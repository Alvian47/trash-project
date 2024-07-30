-- POSTGRES 
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(50) UNIQUE,
    username VARCHAR(30) UNIQUE,
    password VARCHAR(128),
    facebook_id VARCHAR(60),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    title VARCHAR(50),
    content VARCHAR(255),
    status BOOLEAN,
    deadline TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
CREATE TABLE library (
    id SERIAL PRIMARY KEY,
    user_id int,
    task_id int,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (task_id) REFERENCES tasks(id)
);