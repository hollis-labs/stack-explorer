package scip

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/chrispian/stack-explorer/internal/symbols/model"
	scippb "github.com/scip-code/scip/bindings/go/scip"
	"google.golang.org/protobuf/proto"
)

type Ingester struct {
	store            model.Store
	preserveIdentity bool
}

func NewIngester(store model.Store) *Ingester {
	return &Ingester{store: store}
}

func NewAugmentingIngester(store model.Store) *Ingester {
	return &Ingester{store: store, preserveIdentity: true}
}

func (i *Ingester) Ingest(ctx context.Context, req model.IngestRequest) (model.IngestResult, error) {
	if i.store == nil {
		return model.IngestResult{}, fmt.Errorf("scip ingester requires a symbol store")
	}
	languages, err := i.languagesForRequest(req)
	if err != nil {
		return model.IngestResult{}, err
	}
	if len(languages) == 0 {
		if i.preserveIdentity {
			return model.IngestResult{}, nil
		}
		return model.IngestResult{}, fmt.Errorf("no supported SCIP languages detected in %s", req.RepoPath)
	}
	var total model.IngestResult
	for _, language := range languages {
		outPath, err := runIndexer(ctx, language, req)
		if err != nil {
			return total, err
		}
		data, err := os.ReadFile(outPath)
		if err != nil {
			return total, fmt.Errorf("read %s index: %w", language, err)
		}
		result, err := i.ingestIndex(data, req)
		if err != nil {
			return total, fmt.Errorf("ingest %s index: %w", language, err)
		}
		total.Inserted += result.Inserted
		total.Updated += result.Updated
		total.Drifted += result.Drifted
	}
	return total, nil
}

func (i *Ingester) languagesForRequest(req model.IngestRequest) ([]string, error) {
	if len(req.Languages) > 0 {
		languages := normalizeLanguages(req.Languages)
		if i.preserveIdentity {
			return filterIndexableLanguages(languages, req.RepoPath), nil
		}
		if len(languages) == 0 {
			return nil, fmt.Errorf("no supported SCIP languages requested")
		}
		return languages, nil
	}
	languages, err := detectLanguages(req)
	if err != nil {
		if i.preserveIdentity {
			return nil, nil
		}
		return nil, err
	}
	if i.preserveIdentity {
		return filterIndexableLanguages(languages, req.RepoPath), nil
	}
	return languages, nil
}

func (i *Ingester) ingestIndex(data []byte, req model.IngestRequest) (model.IngestResult, error) {
	index := &scippb.Index{}
	if err := proto.Unmarshal(data, index); err != nil {
		return model.IngestResult{}, fmt.Errorf("unmarshal scip index: %w", err)
	}

	parsed, err := collectSymbols(index, req)
	if err != nil {
		return model.IngestResult{}, err
	}

	var result model.IngestResult
	byRawSymbol := map[string]*model.Symbol{}
	parentKeys := map[int64]string{}
	for _, item := range parsed {
		if i.preserveIdentity && shouldSkipAugmentSymbol(item.symbol) {
			continue
		}
		existing, err := i.store.FindSymbolByQualifiedName(req.RepoID, item.symbol.QualifiedName)
		if err != nil {
			return result, err
		}
		if existing != nil {
			item.symbol.ID = existing.ID
			if existing.ContentHash != item.symbol.ContentHash && req.CommitRef != "" {
				item.symbol.StaleSinceCommit = &req.CommitRef
				result.Drifted++
			}
			if i.preserveIdentity {
				if existing.Docstring == "" && item.symbol.Docstring != "" {
					existing.Docstring = item.symbol.Docstring
					if err := i.store.UpsertSymbol(existing); err != nil {
						return result, err
					}
				}
				result.Updated++
				continue
			}
			result.Updated++
		} else {
			if i.preserveIdentity {
				matched, err := i.findExistingByLocation(item.symbol)
				if err != nil {
					return result, err
				}
				if matched != nil {
					if matched.Docstring == "" && item.symbol.Docstring != "" {
						matched.Docstring = item.symbol.Docstring
						if err := i.store.UpsertSymbol(matched); err != nil {
							return result, err
						}
					}
					result.Updated++
					continue
				}
			}
			byHash, err := i.store.FindSymbolByContentHash(req.RepoID, item.symbol.ContentHash)
			if err != nil {
				return result, err
			}
			if byHash != nil {
				if i.preserveIdentity {
					if byHash.Docstring == "" && item.symbol.Docstring != "" {
						byHash.Docstring = item.symbol.Docstring
						if err := i.store.UpsertSymbol(byHash); err != nil {
							return result, err
						}
					}
					result.Updated++
					continue
				}
				item.symbol.ID = byHash.ID
				result.Updated++
			} else {
				result.Inserted++
			}
		}
		if err := i.store.UpsertSymbol(item.symbol); err != nil {
			return result, err
		}
		byRawSymbol[item.rawSymbol] = item.symbol
		if item.parentRawSymbol != "" {
			parentKeys[item.symbol.ID] = item.parentRawSymbol
		}
	}

	for childID, parentRaw := range parentKeys {
		child := byID(byRawSymbol, childID)
		parent := byRawSymbol[parentRaw]
		if child == nil || parent == nil {
			continue
		}
		if child.ParentSymbolID != nil && *child.ParentSymbolID == parent.ID {
			continue
		}
		child.ParentSymbolID = &parent.ID
		if err := i.store.UpsertSymbol(child); err != nil {
			return result, err
		}
	}

	return result, nil
}

type parsedSymbol struct {
	rawSymbol       string
	parentRawSymbol string
	symbol          *model.Symbol
}

func collectSymbols(index *scippb.Index, req model.IngestRequest) ([]parsedSymbol, error) {
	fileCache := map[string][]string{}
	var out []parsedSymbol
	for _, doc := range index.GetDocuments() {
		definitions := definitionRanges(doc)
		for _, info := range doc.GetSymbols() {
			if info.GetSymbol() == "" {
				continue
			}
			sym, err := toStackSymbol(info, doc, definitions[info.GetSymbol()], req, fileCache)
			if err != nil {
				return nil, err
			}
			out = append(out, sym)
		}
	}
	return out, nil
}

type sourceRange struct {
	startLine int
	endLine   int
}

func definitionRanges(doc *scippb.Document) map[string]sourceRange {
	out := map[string]sourceRange{}
	for _, occ := range doc.GetOccurrences() {
		if occ.GetSymbol() == "" {
			continue
		}
		if !scippb.SymbolRole_Definition.Matches(occ) && !scippb.SymbolRole_ForwardDefinition.Matches(occ) {
			continue
		}
		rng := occ.GetEnclosingRange()
		if len(rng) == 0 {
			rng = occ.GetRange()
		}
		if len(rng) < 3 {
			continue
		}
		endLine := int(rng[0])
		if len(rng) >= 4 {
			endLine = int(rng[2])
		}
		out[occ.GetSymbol()] = sourceRange{
			startLine: int(rng[0]),
			endLine:   endLine,
		}
	}
	return out
}

func toStackSymbol(info *scippb.SymbolInformation, doc *scippb.Document, rng sourceRange, req model.IngestRequest, fileCache map[string][]string) (parsedSymbol, error) {
	parsed, err := scippb.ParseSymbol(info.GetSymbol())
	if err != nil {
		return parsedSymbol{}, fmt.Errorf("parse scip symbol %q: %w", info.GetSymbol(), err)
	}

	name := info.GetDisplayName()
	if name == "" && len(parsed.GetDescriptors()) > 0 {
		last := parsed.GetDescriptors()[len(parsed.GetDescriptors())-1]
		name = last.GetName()
	}
	if name == "" {
		name = info.GetSymbol()
	}

	descriptorOnly := scippb.DescriptorOnlyFormatter.FormatSymbol(parsed)
	packageName := ""
	if parsed.GetPackage() != nil {
		packageName = parsed.GetPackage().GetName()
	}
	qualifiedName := descriptorOnly
	if packageName != "" {
		qualifiedName = packageName + ":" + descriptorOnly
	}

	contentHash, signatureHash, err := hashForRange(req.RepoPath, doc.GetRelativePath(), rng, info, fileCache)
	if err != nil {
		return parsedSymbol{}, err
	}

	parentRaw := info.GetEnclosingSymbol()
	if parentRaw == "" && len(parsed.GetDescriptors()) > 1 {
		parent := &scippb.Symbol{
			Scheme:      parsed.GetScheme(),
			Package:     parsed.GetPackage(),
			Descriptors: parsed.GetDescriptors()[:len(parsed.GetDescriptors())-1],
		}
		parentRaw = scippb.VerboseSymbolFormatter.FormatSymbol(parent)
	}

	var lineStart, lineEnd *int
	if rng.startLine >= 0 {
		start := rng.startLine + 1
		end := rng.endLine + 1
		lineStart = &start
		lineEnd = &end
	}

	docstring := strings.TrimSpace(strings.Join(info.GetDocumentation(), "\n\n"))
	return parsedSymbol{
		rawSymbol:       info.GetSymbol(),
		parentRawSymbol: parentRaw,
		symbol: &model.Symbol{
			RepoID:        req.RepoID,
			Kind:          mapKind(info.GetKind()),
			Name:          name,
			QualifiedName: qualifiedName,
			FilePath:      doc.GetRelativePath(),
			LineStart:     lineStart,
			LineEnd:       lineEnd,
			ContentHash:   contentHash,
			SignatureHash: signatureHash,
			Language:      doc.GetLanguage(),
			Visibility:    inferVisibility(name),
			Docstring:     docstring,
		},
	}, nil
}

func hashForRange(repoPath, relativePath string, rng sourceRange, info *scippb.SymbolInformation, fileCache map[string][]string) (string, string, error) {
	snippet, err := fileSnippet(repoPath, relativePath, rng, fileCache)
	if err != nil {
		return "", "", err
	}
	if snippet == "" {
		snippet = info.GetSymbol()
	}
	signature := strings.TrimSpace(info.GetSignatureDocumentation().GetText())
	if signature == "" {
		signature = firstNonEmptyLine(snippet)
	}
	if signature == "" {
		signature = info.GetDisplayName()
	}
	return hashText(snippet), hashText(signature), nil
}

func fileSnippet(repoPath, relativePath string, rng sourceRange, fileCache map[string][]string) (string, error) {
	if rng.startLine < 0 || relativePath == "" {
		return "", nil
	}
	lines, ok := fileCache[relativePath]
	if !ok {
		path := filepath.Join(repoPath, relativePath)
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				return "", nil
			}
			return "", fmt.Errorf("read source file %s: %w", relativePath, err)
		}
		lines = strings.Split(string(data), "\n")
		fileCache[relativePath] = lines
	}
	if rng.startLine >= len(lines) {
		return "", nil
	}
	end := rng.endLine
	if end < rng.startLine {
		end = rng.startLine
	}
	if end >= len(lines) {
		end = len(lines) - 1
	}
	var picked []string
	for idx := rng.startLine; idx <= end; idx++ {
		picked = append(picked, strings.TrimRight(lines[idx], " \t\r"))
	}
	return strings.TrimSpace(strings.Join(picked, "\n")), nil
}

func hashText(text string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(text)))
	return hex.EncodeToString(sum[:])
}

func firstNonEmptyLine(text string) string {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}

func mapKind(kind scippb.SymbolInformation_Kind) string {
	switch kind {
	case scippb.SymbolInformation_Function:
		return "function"
	case scippb.SymbolInformation_Method, scippb.SymbolInformation_MethodAlias, scippb.SymbolInformation_MethodReceiver, scippb.SymbolInformation_MethodSpecification:
		return "method"
	case scippb.SymbolInformation_Class, scippb.SymbolInformation_Struct, scippb.SymbolInformation_Interface, scippb.SymbolInformation_Trait, scippb.SymbolInformation_Object:
		return "type"
	case scippb.SymbolInformation_Constant, scippb.SymbolInformation_EnumMember:
		return "const"
	case scippb.SymbolInformation_Variable, scippb.SymbolInformation_Field, scippb.SymbolInformation_Parameter:
		return "var"
	case scippb.SymbolInformation_Module, scippb.SymbolInformation_Namespace:
		return "module"
	case scippb.SymbolInformation_Package:
		return "package"
	default:
		return strings.ToLower(kind.String())
	}
}

func inferVisibility(name string) string {
	for _, r := range name {
		if unicode.IsLetter(r) {
			if unicode.IsUpper(r) {
				return "public"
			}
			return "private"
		}
	}
	return ""
}

func detectLanguages(req model.IngestRequest) ([]string, error) {
	if len(req.Languages) > 0 {
		languages := normalizeLanguages(req.Languages)
		if len(languages) == 0 {
			return nil, fmt.Errorf("no supported SCIP languages requested")
		}
		return languages, nil
	}
	var out []string
	if fileExists(filepath.Join(req.RepoPath, "go.mod")) {
		out = append(out, "go")
	}
	if fileExists(filepath.Join(req.RepoPath, "package.json")) || fileExists(filepath.Join(req.RepoPath, "tsconfig.json")) {
		out = append(out, "typescript")
	}
	if fileExists(filepath.Join(req.RepoPath, "pyproject.toml")) || fileExists(filepath.Join(req.RepoPath, "setup.py")) || fileExists(filepath.Join(req.RepoPath, "requirements.txt")) {
		out = append(out, "python")
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no supported SCIP languages detected in %s", req.RepoPath)
	}
	return out, nil
}

func normalizeLanguages(languages []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, language := range languages {
		switch strings.ToLower(strings.TrimSpace(language)) {
		case "go":
			if _, ok := seen["go"]; !ok {
				seen["go"] = struct{}{}
				out = append(out, "go")
			}
		case "typescript", "ts", "javascript", "js":
			if _, ok := seen["typescript"]; !ok {
				seen["typescript"] = struct{}{}
				out = append(out, "typescript")
			}
		case "python", "py":
			if _, ok := seen["python"]; !ok {
				seen["python"] = struct{}{}
				out = append(out, "python")
			}
		}
	}
	return out
}

func filterIndexableLanguages(languages []string, repoPath string) []string {
	var out []string
	for _, language := range languages {
		switch language {
		case "go":
			if fileExists(filepath.Join(repoPath, "go.mod")) {
				out = append(out, language)
			}
		case "typescript":
			if fileExists(filepath.Join(repoPath, "tsconfig.json")) {
				out = append(out, language)
			}
		case "python":
			if fileExists(filepath.Join(repoPath, "pyproject.toml")) || fileExists(filepath.Join(repoPath, "setup.py")) || fileExists(filepath.Join(repoPath, "requirements.txt")) {
				out = append(out, language)
			}
		}
	}
	return out
}

func runIndexer(ctx context.Context, language string, req model.IngestRequest) (string, error) {
	tmpDir, err := os.MkdirTemp("", "stack-explorer-scip-*")
	if err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}
	outPath := filepath.Join(tmpDir, language+".scip")
	var name string
	var args []string
	switch language {
	case "go":
		name = "scip-go"
		args = []string{"index", "--module-root", req.RepoPath, "--output", outPath, "./..."}
	case "typescript", "ts", "javascript", "js":
		name = "scip-typescript"
		args = []string{"index", "--cwd", req.RepoPath, "--output", outPath}
	case "python", "py":
		name = "scip-python"
		args = []string{"index", "--cwd", req.RepoPath, "--project-name", req.RepoID, "--output", outPath}
	default:
		return "", fmt.Errorf("unsupported SCIP language %q", language)
	}

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = req.RepoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s failed: %w: %s", name, err, strings.TrimSpace(string(output)))
	}
	return outPath, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func byID(items map[string]*model.Symbol, id int64) *model.Symbol {
	for _, item := range items {
		if item.ID == id {
			return item
		}
	}
	return nil
}

func (i *Ingester) findExistingByLocation(sym *model.Symbol) (*model.Symbol, error) {
	query := normalizeSymbolName(sym.Name)
	if query == "" {
		query = sym.Name
	}
	items, err := i.store.SearchSymbols(model.SearchFilter{
		RepoID: sym.RepoID,
		Query:  query,
		Limit:  50,
	})
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if item.FilePath != sym.FilePath {
			continue
		}
		if !sameSymbolName(item.Name, sym.Name) {
			continue
		}
		if sym.Kind != "" && sym.Kind != "unspecifiedkind" && item.Kind != sym.Kind {
			continue
		}
		if item.LineStart != nil && sym.LineStart != nil && *item.LineStart == *sym.LineStart {
			candidate := item
			return &candidate, nil
		}
	}
	return nil, nil
}

func sameSymbolName(left, right string) bool {
	return normalizeSymbolName(left) == normalizeSymbolName(right)
}

func normalizeSymbolName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.TrimSuffix(name, "()")
	return strings.TrimSpace(name)
}

func shouldSkipAugmentSymbol(sym *model.Symbol) bool {
	return sym.Kind == "unspecifiedkind" && strings.HasSuffix(sym.QualifiedName, "/")
}
