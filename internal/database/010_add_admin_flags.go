package database

func init() {
	RegisterMigration(&Migration{
		ID:          "010_add_admin_flags",
		Description: "Add is_admin and is_super_admin columns to users table",
		Up: map[string]string{
			"sqlite": `
-- Add admin flag columns
ALTER TABLE users ADD COLUMN is_admin INTEGER DEFAULT 0;
ALTER TABLE users ADD COLUMN is_super_admin INTEGER DEFAULT 0;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_admin ON users(is_admin);
CREATE INDEX IF NOT EXISTS idx_users_super_admin ON users(is_super_admin);

-- Create unique constraint to ensure only one super admin exists
CREATE UNIQUE INDEX IF NOT EXISTS idx_one_super_admin ON users(is_super_admin) WHERE is_super_admin = 1;
`,
			"mysql": `
-- Add admin flag columns
ALTER TABLE users ADD COLUMN is_admin TINYINT(1) DEFAULT 0;
ALTER TABLE users ADD COLUMN is_super_admin TINYINT(1) DEFAULT 0;

-- Create indexes for performance
CREATE INDEX idx_users_admin ON users(is_admin);
CREATE INDEX idx_users_super_admin ON users(is_super_admin);

-- Note: MySQL doesn't support partial unique indexes
-- The unique super admin constraint is enforced in application logic
`,
			"postgres": `
-- Add admin flag columns
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_admin BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_super_admin BOOLEAN DEFAULT FALSE;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_admin ON users(is_admin);
CREATE INDEX IF NOT EXISTS idx_users_super_admin ON users(is_super_admin);

-- Create unique constraint to ensure only one super admin exists
CREATE UNIQUE INDEX IF NOT EXISTS idx_one_super_admin ON users(is_super_admin) WHERE is_super_admin = TRUE;
`,
		},
	})
}

// DONE
