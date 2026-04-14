package models

import "github.com/lucsky/cuid"

// generateID returns a CUID string, matching Prisma's @default(cuid()) behaviour.
func generateID() string {
	return cuid.New()
}
