# You are important.

Keep your Slack status and ego in check! Hooks up your flights in Tripit to your Slack status.

# Getting Started

NOTE: Since this is in beta, you'll need to request developer access from TripIt to receive an API
key. Follow the instructions on [this page](https://www.tripit.com/developer) for more info.

1. Add the "Got Status" plugin to TripIt in your Slack Integrations.
2. Type `/givemestatus` to `#slackbot`. It will ask you for your TripIt API key.
3. You should see your status automatically update within a few seconds!

# Privacy, Please!

You might not want a status for all of the trips in your TripIt account. To tell `Status` to
only track specific trips, use the `/givemestatus but only for "<pattern>"` command.

`<pattern>` accepts PCRE regular expressions, but don't worry; `status` will let you know when your
regex is invalid.

## Example

```
/givemestatus but only for "^Foo\:"

> Got it! I've found a few trips in your TripIt that match this expression. Does this look right?
>
> "Foo: 1 2 3"
> "Foo: 4 5 6"

yup

> Awesome! Your changes have been applied.
```

# Status Downgrade

If you no longer want Status to sync your flights, turn it off with `/statusdowngrade`.

# Developing Status

Thanks for your help! Information about developing against Status is forthcoming.

# Want more Status?

Are there any features that you'd like to see? [Submit an
issue](https://github.com/carlosonunez/status/issues/new) and ask for a feature request!
