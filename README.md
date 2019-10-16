# github-clone

> Clone all repos in a GitHub organization organization

Ever moved to a new company and had to clone a bunch of new repos from the private organization? This one hopefully solves it.

## Build
```bash
go build .
```

## Usage
```bash
./github-clone --token "some-token" --org "some-organization" --desination "some-folder"
```

If you use multiple GitHub accounts and therefore SSH keys, you SSH config, e.g. like this:

```bash
Host github.com
  HostName github.com
  User some-user
  IdentityFile ~/.ssh/some-key

Host github-company
  HostName github.com
  User some-other-user
  IdentityFile ~/.ssh/some-other-key
```

In this case- you can specifiy the ```--host``` option to make sure the ```git clone``` matches your SSH config.

Hence, calling ```./github-clone --token "some-token" --org "some-organization" --desination "some-folder" --host "github-company"``` with result in a URL used for cloning like this: ```git@github-company:some-organization/some-repo.git"

## Contributing
1) Fork it (https://github.com/splieth/github-clone/fork).
1) Create your feature branch (git checkout -b feature/some-feature).
1) Commit your changes (git commit -am 'Add some some-feature').
1) Push to the branch (git push origin feature/some-feature).
1) Create a new PR.
1) Profit.
