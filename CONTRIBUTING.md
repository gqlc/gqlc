# Contributing to gqlc

We welcome your contributions to gqlc. This doc describes the
process to contribute patches to gqlc and the general guidelines we
expect contributors to follow. Special thanks to the Protocol Buffers project
because this contributing guideline is a modified form of theirs.

## Before You Start

We accept patches in the form of github pull requests. If you are new to
github, please read [How to create github pull requests](https://help.github.com/articles/about-pull-requests/)
first.

### Coding Style

Please be neat and concise. Always use `go vet` and `go fmt`.

## Contributing Process

Most pull requests should go to the master branch and the change will be
included in the next major/minor version release (e.g., 3.6.0 release). If you
need to include a bug fix in a patch release (e.g., 3.5.2), make sure it’s
already merged to master, and then create a pull request cherry-picking the
commits from master branch to the release branch (e.g., branch 3.5.x).

For each pull request, a gqlc team member will be assigned to review the
pull request. For minor cleanups, the pull request may be merged right away
after an initial review. For larger changes, you will likely receive multiple
rounds of comments and it may take some time to complete. We will try to keep
our response time within 7-days but if you don’t get any response in a few
days, feel free to comment on the threads to get our attention. We also expect
you to respond to our comments within a reasonable amount of time. If we don’t
hear from you for 2 weeks or longer, we may close the pull request. You can
still send the pull request again once you have time to work on it.

Once a pull request is merged, we will take care of the rest and get it into
the final release.

## Pull Request Guidelines

* Create small PRs that are narrowly focused on addressing a single concern.
  We often receive PRs that are trying to fix several things at a time, but if
  only one fix is considered acceptable, nothing gets merged and both author's
  & review's time is wasted. Create more PRs to address different concerns and
  everyone will be happy.
* For speculative changes, consider opening an issue and discussing it first.
  If you are suggesting a behavioral or API change, make sure you get explicit
  support from a gqlc team member before sending us the pull request.
* Provide a good PR description as a record of what change is being made and
  why it was made. Link to a GitHub issue if it exists.
* Don't fix code style and formatting unless you are already changing that
  line to address an issue. PRs with irrelevant changes won't be merged. If
  you do want to fix formatting or style, do that in a separate PR.
* Unless your PR is trivial, you should expect there will be reviewer comments
  that you'll need to address before merging. We expect you to be reasonably
  responsive to those comments, otherwise the PR will be closed after 2-3 weeks
  of inactivity.
* Maintain clean commit history and use meaningful commit messages. PRs with
  messy commit history are difficult to review and won't be merged. Use rebase
  -i upstream/master to curate your commit history and/or to bring in latest
  changes from master (but avoid rebasing in the middle of a code review).
* Keep your PR up to date with upstream/master (if there are merge conflicts,
  we can't really merge your change).
* All tests need to be passing before your change can be merged. We recommend
  you run tests locally before creating your PR to catch breakages early on.
  Ultimately, the green signal will be provided by our testing infrastructure.
  The reviewer will help you if there are test failures that seem not related
  to the change you are making.
