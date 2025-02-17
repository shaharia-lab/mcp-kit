# MCP Kit (Model Context Protocol) Kit

## Installation

### Using Docker

```bash
# Build the image
docker build -t mcp:latest .

# Run server
docker run -d \
  --name mcp-server \
  -p 8080:8080 \
  -e MCP_SERVER_PORT=8080 \
  mcp:latest server

# Run client
docker run -d \
  --name mcp-client \
  --add-host=host.docker.internal:host-gateway \
  -e MCP_SERVER_URL=http://host.docker.internal:8080/events \
  mcp:latest api
```

### Test Postgresql Tools

```bash
docker-compose up -d
```

```sql
-- Create tables
CREATE TABLE IF NOT EXISTS customers (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    username VARCHAR(50) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT true,
    CONSTRAINT uk_customers_email UNIQUE (email),
    CONSTRAINT uk_customers_username UNIQUE (username)
);

CREATE TABLE IF NOT EXISTS profiles (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER NOT NULL,
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    phone VARCHAR(20),
    date_of_birth DATE,
    address TEXT,
    city VARCHAR(100),
    country VARCHAR(100),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_profiles_customer FOREIGN KEY (customer_id) 
        REFERENCES customers(id) ON DELETE CASCADE,
    CONSTRAINT uk_profiles_customer UNIQUE (customer_id)
);

-- Insert dummy data
INSERT INTO customers (email, username, password_hash, last_login, is_active)
SELECT 
    'user' || n || '@example.com',
    'user' || n,
    'hashed_password_' || MD5(n::text),
    CURRENT_TIMESTAMP - (n || ' days')::INTERVAL,
    n % 5 != 0  -- Every 5th user is inactive
FROM generate_series(1, 10) n
WHERE NOT EXISTS (
    SELECT 1 FROM customers 
    WHERE email = 'user' || n || '@example.com'
);

INSERT INTO profiles (customer_id, first_name, last_name, phone, date_of_birth, address, city, country)
SELECT 
    c.id,
    CASE (n % 3)
        WHEN 0 THEN 'John'
        WHEN 1 THEN 'Jane'
        ELSE 'Alex'
    END,
    CASE (n % 3)
        WHEN 0 THEN 'Doe'
        WHEN 1 THEN 'Smith'
        ELSE 'Johnson'
    END,
    '+1-555-' || LPAD(n::text, 4, '0'),
    '1990-01-01'::DATE + (n * 100 || ' days')::INTERVAL,
    n || ' Main Street',
    CASE (n % 3)
        WHEN 0 THEN 'New York'
        WHEN 1 THEN 'Los Angeles'
        ELSE 'Chicago'
    END,
    'USA'
FROM customers c
CROSS JOIN generate_series(1, 10) n
WHERE c.id = n
AND NOT EXISTS (
    SELECT 1 FROM profiles 
    WHERE customer_id = c.id
);

-- Add some indexes for better performance
CREATE INDEX IF NOT EXISTS idx_customers_email ON customers(email);
CREATE INDEX IF NOT EXISTS idx_customers_username ON customers(username);
CREATE INDEX IF NOT EXISTS idx_profiles_customer_id ON profiles(customer_id);
CREATE INDEX IF NOT EXISTS idx_profiles_city_country ON profiles(city, country);
```