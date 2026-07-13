-- Create a dedicated namespace and table for FT-13 negative test (irreversible type narrowing).
-- Uses isolated namespace "raw_ft13" to avoid conflicts with the main fixture chain.
CREATE NAMESPACE raw_ft13;
CREATE TABLE iceberg.raw_ft13.irreversible_demo (id int);
