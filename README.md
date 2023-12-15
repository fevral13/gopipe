# about

This is a simple console application that I created for myself to monitor our GitLab pipelines. 
Every `delay` seconds it fetches pipelines metadata of main branches (configurable) and branches of all open MRs.

<img width="1285" alt="Screenshot 2023-12-15 at 12 45 47" src="https://github.com/fevral13/gopipe/assets/183431/ed37104c-e9d9-4da1-8075-ae19bbb15d8d">

# usage

You need go compiler to build the app.

    $ brew install go

Clone this repo, then run

    $ go build

I recommend to install gopipe binary with `go install` in your home dir: `~/go/bin/gopipe`. Make sure that `~/go/bin` is in your `PATH`.

gopipe gets its configuration from your ENV. These vars are available:

```
# optional, defaults to https://gitlab.com/
GP_URL=

# Personal Access Token. required, get one in your settings <gitlab url>/-/profile/personal_access_tokens. It requires only Read permissions
GP_API_KEY=glpat-.....

# project ID. required
GP_PROJECT_ID=56

# update interval, in seconds. Optional, default is 10
GP_DELAY=10

# comma-separated list of branches, pipelines of which you want to monitor, in addition to open MRs. Optional
GP_MAIN_BRANCHES=development,release-23.12.x,release-23.11.x,release-23.10.x

```

Export these variables, or put them in your `.bashrc/.zshrc/...` and run `gopipe`.

Usually I have it running in a separate small terminal window.
