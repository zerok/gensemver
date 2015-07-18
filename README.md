# Generate Semantic Version Numbers

This is a simple tool that parses your git commit history (of the master branch
only) in order to determine the next version number you should tag your
repository with. The idea here is that once you merge something into the master
branch, the difference between the state before the merge and after it is parsed
to determine if this is a feature, patch, or major release based on the commits
made between the current and the previous state.

```
$ gensemver
2015/07/18 21:13:45 Previous version: 1.0.0 (0930a6e33deb0068d14bc031f8eb6287f6f1a98d)
1.1.0
```

Not passing any parameters will make the tool look into the tag history and work
based on the latest tag up to HEAD.

You can also just pass a commit range to the tool in order to get a good idea of
what to expect.

```
$ gensemver HEAD^^ HEAD
2015/07/18 21:13:45 Previous version: 1.0.0 (0930a6e33deb0068d14bc031f8eb6287f6f1a98d)
1.1.0
```

If you only pass one positional parameter, the second is implied to be HEAD.

```
$ gensemver HEAD^^
2015/07/18 21:13:45 Previous version: 1.0.0 (0930a6e33deb0068d14bc031f8eb6287f6f1a98d)
1.1.0
```

If you want to force a starting version number (for example because you don't
encoding version numbers in release tags for some reason) you can specify one
with the `-prev=<version>` flag.


## Commit message format

The commit messages have to be structured in a way similar to how the AngularJS
project handled them. The general format is like this:

```
<type>[(<component>)]: <short info>

[<details>]
```

The `<type>` can be one of the following:

* chore
* doc
* test
* feat
* fix

The optional component should indicate what part of the project was changed. In
general it helps to keep the `<short info>` short.

The type combined with some information from the `<details>` section defines the
status of the next release:

* chore won't influence the version number
* doc, test, fix will suggest a patch release
* feat a minor release
* any of the above with the text "BREAKING CHANGES:" in the details section
  forces a major release.

This means that much care has to go into the creation of these commit
messages. As such it is recommended that you use a pre-commit hook to enforce
this format.
