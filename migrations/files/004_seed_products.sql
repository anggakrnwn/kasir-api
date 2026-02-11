
INSERT INTO products (name, price, stock) VALUES
    ('Indomie Goreng', 3500, 100),
    ('Indomie Rebus', 3500, 100),
    ('Aqua 600ml', 2000, 200),
    ('Kopi Kapal Api', 2500, 150),
    ('Roti Tawar', 12000, 30)
ON CONFLICT DO NOTHING;