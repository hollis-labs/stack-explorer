INSERT OR IGNORE INTO review_dimensions (id, name, category, description, weight, sort_order) VALUES
    ('hooks',              'Hooks',                'agent_patterns', 'Lifecycle hooks, enforcement hooks, extension hooks', 1.0, 1),
    ('loops',              'Loops',                'agent_patterns', 'Agent loops, feedback loops, retry patterns', 1.0, 2),
    ('tool_calls',         'Tool Calls',           'agent_patterns', 'Security, permissions, long-running, token efficiency', 1.0, 3),
    ('agents',             'Agents',               'agent_patterns', 'Agent definition, coordination, delegation patterns', 1.0, 4),
    ('skills',             'Skills',               'agent_patterns', 'Skill discovery, composition, reuse', 1.0, 5),
    ('progressive_context','Progressive Context',  'agent_patterns', 'Progressive context discovery and loading', 1.0, 6),
    ('memory',             'Memory',               'agent_patterns', 'Persistence, retrieval, decay patterns', 1.0, 7),
    ('agent_coordination', 'Agent Coordination',   'agent_patterns', 'Multi-agent communication and handoff', 1.0, 8),
    ('stack',              'Stack Details',         'code_quality',   'Language, framework, dependency choices', 0.5, 9),
    ('loc',                'Codebase Size',         'code_quality',   'Lines of code, file count, growth rate', 0.5, 10),
    ('security',           'Security Posture',      'project_health', 'Dependency scanning, secret management, input validation', 1.0, 11),
    ('maturity',           'Maturity',              'project_health', 'Tests, docs, CI/CD, release cadence', 1.0, 12),
    ('complexity',         'Complexity',            'code_quality',   'Cyclomatic, cognitive, architectural complexity', 1.0, 13);
