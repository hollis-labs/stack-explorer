CREATE VIRTUAL TABLE IF NOT EXISTS findings_fts USING fts5(
    title,
    description,
    body_markdown,
    content='findings',
    content_rowid='id'
);

CREATE VIRTUAL TABLE IF NOT EXISTS symbols_fts USING fts5(
    name,
    qualified_name,
    file_path,
    docstring,
    content='symbols',
    content_rowid='id'
);

CREATE VIRTUAL TABLE IF NOT EXISTS code_references_fts USING fts5(
    file_path,
    description,
    content='code_references',
    content_rowid='id'
);

INSERT INTO findings_fts(rowid, title, description, body_markdown)
SELECT id, title, description, body_markdown FROM findings;

INSERT INTO symbols_fts(rowid, name, qualified_name, file_path, docstring)
SELECT id, name, qualified_name, file_path, docstring FROM symbols;

INSERT INTO code_references_fts(rowid, file_path, description)
SELECT id, file_path, description FROM code_references;

CREATE TRIGGER IF NOT EXISTS findings_fts_ai AFTER INSERT ON findings BEGIN
    INSERT INTO findings_fts(rowid, title, description, body_markdown)
    VALUES (new.id, new.title, new.description, new.body_markdown);
END;

CREATE TRIGGER IF NOT EXISTS findings_fts_ad AFTER DELETE ON findings BEGIN
    INSERT INTO findings_fts(findings_fts, rowid, title, description, body_markdown)
    VALUES ('delete', old.id, old.title, old.description, old.body_markdown);
END;

CREATE TRIGGER IF NOT EXISTS findings_fts_au AFTER UPDATE ON findings BEGIN
    INSERT INTO findings_fts(findings_fts, rowid, title, description, body_markdown)
    VALUES ('delete', old.id, old.title, old.description, old.body_markdown);
    INSERT INTO findings_fts(rowid, title, description, body_markdown)
    VALUES (new.id, new.title, new.description, new.body_markdown);
END;

CREATE TRIGGER IF NOT EXISTS symbols_fts_ai AFTER INSERT ON symbols BEGIN
    INSERT INTO symbols_fts(rowid, name, qualified_name, file_path, docstring)
    VALUES (new.id, new.name, new.qualified_name, new.file_path, new.docstring);
END;

CREATE TRIGGER IF NOT EXISTS symbols_fts_ad AFTER DELETE ON symbols BEGIN
    INSERT INTO symbols_fts(symbols_fts, rowid, name, qualified_name, file_path, docstring)
    VALUES ('delete', old.id, old.name, old.qualified_name, old.file_path, old.docstring);
END;

CREATE TRIGGER IF NOT EXISTS symbols_fts_au AFTER UPDATE ON symbols BEGIN
    INSERT INTO symbols_fts(symbols_fts, rowid, name, qualified_name, file_path, docstring)
    VALUES ('delete', old.id, old.name, old.qualified_name, old.file_path, old.docstring);
    INSERT INTO symbols_fts(rowid, name, qualified_name, file_path, docstring)
    VALUES (new.id, new.name, new.qualified_name, new.file_path, new.docstring);
END;

CREATE TRIGGER IF NOT EXISTS code_references_fts_ai AFTER INSERT ON code_references BEGIN
    INSERT INTO code_references_fts(rowid, file_path, description)
    VALUES (new.id, new.file_path, new.description);
END;

CREATE TRIGGER IF NOT EXISTS code_references_fts_ad AFTER DELETE ON code_references BEGIN
    INSERT INTO code_references_fts(code_references_fts, rowid, file_path, description)
    VALUES ('delete', old.id, old.file_path, old.description);
END;

CREATE TRIGGER IF NOT EXISTS code_references_fts_au AFTER UPDATE ON code_references BEGIN
    INSERT INTO code_references_fts(code_references_fts, rowid, file_path, description)
    VALUES ('delete', old.id, old.file_path, old.description);
    INSERT INTO code_references_fts(rowid, file_path, description)
    VALUES (new.id, new.file_path, new.description);
END;
