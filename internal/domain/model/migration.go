package model

type Migration struct {
	Version     string
	ApplyTime   int
	BodySQL     string
	ExecutedSQL string
	Release     string
}
type Migrations []Migration
