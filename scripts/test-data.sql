-- Insert test entities for TUI testing
INSERT INTO discovered_entities (unique_id, type_category, name, properties, confidence_score, created_at)
SELECT 
  'person:' || lower(replace(name, ' ', '.')),
  'person',
  name,
  '{}'::jsonb,
  1.0,
  NOW()
FROM (VALUES ('Greg Whalley'), ('Rebecca Mark'), ('Kenneth Lay'), ('Jeff Skilling')) AS t(name)
UNION ALL
SELECT 
  'org:' || lower(replace(name, ' ', '-')),
  'organization',
  name,
  '{}'::jsonb,
  0.9,
  NOW()
FROM (VALUES ('Enron'), ('West Power Trading Desk'), ('East Power Trading Desk')) AS t(name)
UNION ALL
SELECT 
  'loc:london',
  'location',
  'London',
  '{}'::jsonb,
  0.8,
  NOW();

-- Insert test relationships
INSERT INTO relationships (type, from_type, from_id, to_type, to_id, confidence_score, timestamp, properties, created_at)
VALUES
  ('WORKS_AT', 'person', 1, 'organization', 5, 0.95, NOW(), '{}', NOW()),
  ('COMMUNICATES_WITH', 'person', 1, 'person', 4, 0.90, NOW(), '{}', NOW()),
  ('MENTIONS', 'person', 1, 'location', 8, 0.80, NOW(), '{}', NOW()),
  ('WORKS_AT', 'person', 2, 'organization', 5, 0.95, NOW(), '{}', NOW()),
  ('LEADS', 'person', 2, 'organization', 6, 0.85, NOW(), '{}', NOW()),
  ('LEADS', 'person', 3, 'organization', 5, 0.98, NOW(), '{}', NOW()),
  ('COMMUNICATES_WITH', 'person', 3, 'person', 4, 0.92, NOW(), '{}', NOW()),
  ('WORKS_AT', 'person', 4, 'organization', 5, 0.95, NOW(), '{}', NOW()),
  ('MANAGES', 'person', 4, 'organization', 7, 0.88, NOW(), '{}', NOW()),
  ('LOCATED_IN', 'organization', 7, 'location', 8, 0.75, NOW(), '{}', NOW());
