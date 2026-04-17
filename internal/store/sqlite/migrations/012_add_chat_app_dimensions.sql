-- New dimensions for chat-app and enterprise lenses
INSERT OR IGNORE INTO review_dimensions (id, name, category, description, weight, sort_order) VALUES
    ('multi_provider',   'Multi-Provider',       'features',    'Support for multiple LLM providers and models', 1.0, 14),
    ('multi_modal',      'Multi-Modal',          'features',    'Support for text, image, audio, video, and other modalities', 1.0, 15),
    ('ui_extensibility', 'UI Extensibility',     'features',    'Widgets, plugins, custom components, envelope/card systems', 1.0, 16),
    ('enterprise',       'Enterprise Readiness',  'project_health', 'SSO, RBAC, audit logging, data governance, compliance controls', 1.0, 17),
    ('data_governance',  'Data Governance',       'project_health', 'Data retention, redaction, PII handling, export controls', 1.0, 18);
