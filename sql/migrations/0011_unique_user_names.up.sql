-- Names remain reserved after soft deletion so restore preserves identity.
-- If duplicates already exist, this migration fails without changing users.
-- Before upgrading, inspect conflicts with:
-- SELECT name, array_agg(id ORDER BY id) FROM anclax.users
-- GROUP BY name HAVING count(*) > 1;
-- Resolve ownership explicitly; never auto-delete or merge these accounts.
ALTER TABLE anclax.users ADD CONSTRAINT users_name_key UNIQUE (name);
