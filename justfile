set dotenv-load := true

module_url := `head -n 1 go.mod | cut -d ' ' -f 2`

test:
	go test ./...

updsum SEMVER:
	sleep 3
	curl https://sum.golang.org/lookup/{{module_url}}@{{SEMVER}}

tag-push SEMVER:
	git tag {{SEMVER}}
	git push origin {{SEMVER}}

commit-push SEMVER:
    git add . ; git commit -m "release: {{SEMVER}}"
    git tag {{SEMVER}}
    git push ; git push origin {{SEMVER}}

release SEMVER: test (commit-push SEMVER) (updsum SEMVER)
