## Notes for Reviewer
<!-- Items in addition to the checklist below that the reviewer should look for -->

## Pre-Review Checklist for the PR Author

* [ ] Add sensible notes for the reviewer
* [ ] PR title is short, expressive and meaningful
* [ ] Consider switching the PR to a draft (`Convert to draft`)
    * as draft PR, the CI will be skipped for pushes
* [ ] Relevant issues are linked in the [References](#references) section
* [ ] Branch follows the naming format (`iox2-123-short-description`)
* [ ] Commits messages are according to this [guideline][commit-guidelines]
    * [ ] Commit messages have the issue ID (`[#123] Add feature description`)
* [ ] Tests have been added/updated for new functionality
* [ ] Exported identifiers are documented
* [ ] `golangci-lint run ./...` passes
* [ ] Assign PR to reviewer
* [ ] All checks have passed

[commit-guidelines]: https://tbaggery.com/2008/04/19/a-note-about-git-commit-messages.html

## PR Reviewer Reminders

* Commits are properly organized and messages are according to the guideline
* Unit tests have been written for new behavior
* Exported API is documented
* PR title describes the changes
* Native handles are released and `Close` is idempotent

## References

<!-- Use either 'Closes #123' or 'Relates to #123' to reference the corresponding issue. -->

Closes # <!-- Add issue number after '#' -->
