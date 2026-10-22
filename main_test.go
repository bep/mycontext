package main

import (
	"os"
	"testing"

	"github.com/bep/helpers/envhelpers"
	"github.com/rogpeppe/go-internal/testscript"
)

func TestScripts(t *testing.T) {
	setup := testSetupFunc()
	testscript.Run(t, testscript.Params{
		Dir: "testscripts",
		// UpdateScripts: true, // Uncomment to rewrite the test scripts with
		// TestWork: true, // Uncomment to keep the test work dir.
		Setup: func(env *testscript.Env) error {
			return setup(env)
		},
	})
}

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"mycontext": main,
	})
}

func testSetupFunc() func(env *testscript.Env) error {
	sourceDir, _ := os.Getwd()
	return func(env *testscript.Env) error {
		envhelpers.SetEnvVars(&env.Vars,
			"SOURCE", sourceDir,
			"GIT_CONFIG_NOSYSTEM", "1",
			"GIT_AUTHOR_NAME", "Test",
			"GIT_AUTHOR_EMAIL", "test@example.com",
			"GIT_COMMITTER_NAME", "Test",
			"GIT_COMMITTER_EMAIL", "test@example.com",
		)
		return nil
	}
}
