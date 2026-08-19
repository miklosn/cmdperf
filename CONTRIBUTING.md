# Contributing to cmdperf

Thank you for considering contributing to `cmdperf`! We appreciate any help you can offer, whether it's code, documentation, bug reports, or feedback.

## Getting Started

1.  **Prerequisites:** Ensure you have [Go](https://golang.org/doc/install) installed (check `go.mod` for the required version).
2.  **Fork & Clone:** Fork the repository and clone your fork locally.
3.  **Build:** Run `make build` to build the project.
4.  **Test:** Run `make test` to execute the test suite.

## Making Changes (Code Contributions)

1.  Create a new branch for your changes.
2.  Make your modifications.
3.  Ensure tests pass (`make test`).
4.  Add an entry under `## [Unreleased]` in `CHANGELOG.md` if your change is
    user-visible. `CHANGELOG.md` is the source of published release notes, so
    anything not written there does not reach users.
5.  Commit your changes with a clear message.
6.  Push your branch to your fork.

## Submitting Contributions

-   **Issues (Bugs, Features, Feedback):** We highly welcome issue reports! Please open an issue on the main repository detailing any bugs you find, feature ideas you have, or general feedback you'd like to share.
-   **Pull Requests (Code/Docs):** For code or documentation changes, open a pull request from your fork's branch to the main repository's `main` branch. Briefly describe the changes you've made.

We aim for a simple and effective contribution process. Thanks again for your interest!

## Releases

Release notes are published from `CHANGELOG.md`, not generated from the commit
log. To cut a release:

1.  Rename `## [Unreleased]` to `## [X.Y.Z] - YYYY-MM-DD` and add a new empty
    `## [Unreleased]` above it, then update the compare links at the bottom.
2.  Preview what will be published: `mise run release-notes vX.Y.Z`.
3.  Tag and push `vX.Y.Z`, or run the Release workflow with that version.

The release job extracts the section matching the tag and passes it to
goreleaser. If that section is missing it falls back to `## [Unreleased]`, and
fails the release if neither has content.
