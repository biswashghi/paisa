CREATE SCHEMA IF NOT EXISTS paisa;

-- Create table accounts with accountid uuid and balance, Generate accountid if not provided and set balance to 0
CREATE TABLE IF NOT EXISTS paisa.accounts (
    accountid UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    balance NUMERIC(10, 2) DEFAULT 0
);

-- Create table transactions with transactionid uuid, accountid uuid, description, merchantcode, balance, and timestamp
CREATE TABLE IF NOT EXISTS paisa.transactions (
    transactionid UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    accountid UUID NOT NULL,
    description VARCHAR(255) NOT NULL,
    merchantcode VARCHAR(255) NOT NULL,
    increment NUMERIC(10, 2) NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (accountid) REFERENCES paisa.accounts(accountid)
);

-- Create table users for authentication
CREATE TABLE IF NOT EXISTS paisa.users (
    accountid UUID PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL
);

-- Create table sessions for session management
CREATE TABLE IF NOT EXISTS paisa.sessions (
    sessionid UUID PRIMARY KEY,
    accountid UUID NOT NULL,
    createdat TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updatedat TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (accountid) REFERENCES paisa.users(accountid)
);