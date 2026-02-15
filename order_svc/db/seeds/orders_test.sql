TRUNCATE TABLE orders;

INSERT INTO orders (id, items, status, created_at, updated_at) VALUES
(
    '11111111-1111-1111-1111-111111111111',
    '{"sku_1": 2, "sku_2": 1}'::jsonb,
    'pending',
    NOW() - INTERVAL '10 minutes',
    NOW() - INTERVAL '10 minutes'
),
(
    '22222222-2222-2222-2222-222222222222',
    '{"sku_3": 5}'::jsonb,
    'confirmed',
    NOW() - INTERVAL '5 minutes',
    NOW() - INTERVAL '5 minutes'
),
(
    '33333333-3333-3333-3333-333333333333',
    '{"sku_4": 1, "sku_5": 3}'::jsonb,
    'failed',
    NOW() - INTERVAL '1 minute',
    NOW() - INTERVAL '1 minute'
);
