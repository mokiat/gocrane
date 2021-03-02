package command

import "github.com/mokiat/gocrane/internal/project"

func addDefaultExcludes(current []string) []string {
	return append(current,
		project.Glob(".git"),
		project.Glob(".github"),
		project.Glob(".gitignore"),
		project.Glob(".DS_Store"),
		project.Glob(".vscode"),
		project.Glob(".md"),
		project.Glob("Dockerfile"),
		project.Glob("LICENSE"),
		project.Glob("*_test.go"),
	)
}

func addDefaultResources(current []string) []string {
	return append(current,
		project.Glob(".json"),
		project.Glob(".xml"),
		project.Glob(".yml"),
		project.Glob(".yaml"),
		project.Glob(".html"),
		project.Glob(".js"),
		project.Glob(".png"),
		project.Glob(".jpg"),
		project.Glob(".jpeg"),
		project.Glob(".css"),
	)
}
