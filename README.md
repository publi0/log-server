# Log Server

This is a simple log server that listens for logs and if matches a certain pattern, it will send an notification to ntfy.

## Create table

```sql
CREATE TABLE network_status (
    id SERIAL PRIMARY KEY,
    wan VARCHAR(255),
    status VARCHAR(255),
    create_at TIMESTAMP WITH TIME ZONE
);
```
