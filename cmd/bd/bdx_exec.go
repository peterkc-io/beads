package main

import "github.com/steveyegge/beads/internal/utils"

// selfCommand lets hook tests substitute a subprocess without re-executing tests.
var selfCommand = utils.SelfCommand
