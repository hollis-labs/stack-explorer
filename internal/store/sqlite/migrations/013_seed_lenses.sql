-- Default lenses
INSERT OR IGNORE INTO lenses (id, name, description, created_at) VALUES
    ('agent-platform',   'Agent Platform',      'Full agent platform with tool calling, coordination, skills',        '2026-04-03T00:00:00Z'),
    ('chat-app',         'Chat Application',    'GUI AI chat apps — tool calls, security, enterprise, features, UX',  '2026-04-03T00:00:00Z'),
    ('automation',       'Automation / DAG',    'Workflow automation, DAG orchestration, scheduling',                  '2026-04-03T00:00:00Z'),
    ('memory-system',    'Memory System',       'Memory, RAG, embeddings, context management',                        '2026-04-03T00:00:00Z'),
    ('agent-framework',  'Agent Framework',     'Agent definition, skills, coordination, tool calling',               '2026-04-03T00:00:00Z'),
    ('desktop-app',      'Desktop Application', 'Desktop/GUI apps — UX, stability, simplicity',                       '2026-04-03T00:00:00Z'),
    ('infra-tool',       'Infrastructure Tool', 'Service management, monitoring, ops tooling',                         '2026-04-03T00:00:00Z'),
    ('content-pipeline', 'Content Pipeline',    'Ingest, transform, and output content',                              '2026-04-03T00:00:00Z'),
    ('general',          'General',             'All dimensions equally weighted — default lens',                      '2026-04-03T00:00:00Z');

-- agent-platform: all original 13 dimensions
INSERT OR IGNORE INTO lens_dimensions (lens_id, dimension_id, weight, sort_order) VALUES
    ('agent-platform', 'hooks', 1.0, 1),
    ('agent-platform', 'loops', 1.0, 2),
    ('agent-platform', 'tool_calls', 1.0, 3),
    ('agent-platform', 'agents', 1.0, 4),
    ('agent-platform', 'skills', 1.0, 5),
    ('agent-platform', 'progressive_context', 1.0, 6),
    ('agent-platform', 'memory', 1.0, 7),
    ('agent-platform', 'agent_coordination', 1.0, 8),
    ('agent-platform', 'stack', 0.5, 9),
    ('agent-platform', 'loc', 0.5, 10),
    ('agent-platform', 'security', 1.0, 11),
    ('agent-platform', 'maturity', 1.0, 12),
    ('agent-platform', 'complexity', 1.0, 13);

-- chat-app: tool calls, security, enterprise, features, plus foundational
INSERT OR IGNORE INTO lens_dimensions (lens_id, dimension_id, weight, sort_order) VALUES
    ('chat-app', 'tool_calls', 1.5, 1),
    ('chat-app', 'security', 1.5, 2),
    ('chat-app', 'enterprise', 1.5, 3),
    ('chat-app', 'data_governance', 1.0, 4),
    ('chat-app', 'multi_provider', 1.5, 5),
    ('chat-app', 'multi_modal', 1.0, 6),
    ('chat-app', 'ui_extensibility', 1.5, 7),
    ('chat-app', 'memory', 1.0, 8),
    ('chat-app', 'hooks', 0.5, 9),
    ('chat-app', 'agents', 0.5, 10),
    ('chat-app', 'skills', 0.5, 11),
    ('chat-app', 'stack', 0.5, 12),
    ('chat-app', 'maturity', 1.0, 13),
    ('chat-app', 'complexity', 0.5, 14);

-- automation: hooks, loops, skills, memory, tool_calls, security, maturity, complexity, stack, loc
INSERT OR IGNORE INTO lens_dimensions (lens_id, dimension_id, weight, sort_order) VALUES
    ('automation', 'hooks', 1.0, 1),
    ('automation', 'loops', 1.5, 2),
    ('automation', 'skills', 1.5, 3),
    ('automation', 'tool_calls', 1.0, 4),
    ('automation', 'memory', 1.0, 5),
    ('automation', 'security', 1.0, 6),
    ('automation', 'maturity', 1.0, 7),
    ('automation', 'complexity', 1.0, 8),
    ('automation', 'stack', 0.5, 9),
    ('automation', 'loc', 0.5, 10);

-- memory-system
INSERT OR IGNORE INTO lens_dimensions (lens_id, dimension_id, weight, sort_order) VALUES
    ('memory-system', 'memory', 2.0, 1),
    ('memory-system', 'progressive_context', 1.5, 2),
    ('memory-system', 'tool_calls', 1.0, 3),
    ('memory-system', 'security', 1.0, 4),
    ('memory-system', 'maturity', 1.0, 5),
    ('memory-system', 'complexity', 1.0, 6),
    ('memory-system', 'stack', 0.5, 7),
    ('memory-system', 'loc', 0.5, 8);

-- agent-framework
INSERT OR IGNORE INTO lens_dimensions (lens_id, dimension_id, weight, sort_order) VALUES
    ('agent-framework', 'agents', 1.5, 1),
    ('agent-framework', 'skills', 1.5, 2),
    ('agent-framework', 'hooks', 1.0, 3),
    ('agent-framework', 'loops', 1.0, 4),
    ('agent-framework', 'tool_calls', 1.5, 5),
    ('agent-framework', 'agent_coordination', 1.5, 6),
    ('agent-framework', 'maturity', 1.0, 7),
    ('agent-framework', 'complexity', 1.0, 8),
    ('agent-framework', 'stack', 0.5, 9);

-- desktop-app
INSERT OR IGNORE INTO lens_dimensions (lens_id, dimension_id, weight, sort_order) VALUES
    ('desktop-app', 'memory', 1.0, 1),
    ('desktop-app', 'security', 1.0, 2),
    ('desktop-app', 'maturity', 1.0, 3),
    ('desktop-app', 'complexity', 1.0, 4),
    ('desktop-app', 'stack', 1.0, 5),
    ('desktop-app', 'loc', 1.0, 6);

-- infra-tool
INSERT OR IGNORE INTO lens_dimensions (lens_id, dimension_id, weight, sort_order) VALUES
    ('infra-tool', 'hooks', 1.0, 1),
    ('infra-tool', 'loops', 1.0, 2),
    ('infra-tool', 'security', 1.5, 3),
    ('infra-tool', 'maturity', 1.0, 4),
    ('infra-tool', 'complexity', 1.0, 5),
    ('infra-tool', 'stack', 0.5, 6),
    ('infra-tool', 'loc', 0.5, 7);

-- content-pipeline
INSERT OR IGNORE INTO lens_dimensions (lens_id, dimension_id, weight, sort_order) VALUES
    ('content-pipeline', 'hooks', 1.0, 1),
    ('content-pipeline', 'loops', 1.0, 2),
    ('content-pipeline', 'memory', 1.0, 3),
    ('content-pipeline', 'security', 1.0, 4),
    ('content-pipeline', 'maturity', 1.0, 5),
    ('content-pipeline', 'complexity', 1.0, 6),
    ('content-pipeline', 'stack', 0.5, 7),
    ('content-pipeline', 'loc', 0.5, 8);

-- general: all 18 dimensions equally weighted
INSERT OR IGNORE INTO lens_dimensions (lens_id, dimension_id, weight, sort_order) VALUES
    ('general', 'hooks', 1.0, 1),
    ('general', 'loops', 1.0, 2),
    ('general', 'tool_calls', 1.0, 3),
    ('general', 'agents', 1.0, 4),
    ('general', 'skills', 1.0, 5),
    ('general', 'progressive_context', 1.0, 6),
    ('general', 'memory', 1.0, 7),
    ('general', 'agent_coordination', 1.0, 8),
    ('general', 'stack', 1.0, 9),
    ('general', 'loc', 1.0, 10),
    ('general', 'security', 1.0, 11),
    ('general', 'maturity', 1.0, 12),
    ('general', 'complexity', 1.0, 13),
    ('general', 'multi_provider', 1.0, 14),
    ('general', 'multi_modal', 1.0, 15),
    ('general', 'ui_extensibility', 1.0, 16),
    ('general', 'enterprise', 1.0, 17),
    ('general', 'data_governance', 1.0, 18);
