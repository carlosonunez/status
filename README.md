# Status Bot

A small command-line tool that keeps your status in sync with your whereabouts.

## Examples

- Set your Slack status to "✈️  in flight!" during the flight portion of an
  active trip on TripIt
- Tell people on WhatsApp that you're "🚗 blasting Nirvana" when your iPhone goes
  into Driving mode (through an iOS Shortcut)
- Inform developers on yoru team with an "🌴 Vacationing" status on GitHub when
  you add an all-day "Out of Office" event into Google Calendar

## You'll Like This If

- You travel a lot and want to keep people informed on your whereabouts without
  having to dig through menus every time.
- You want more control over your notifications without relying on a million
  different tools and features.

## Integratations

> 📝 You can create your own integrations! Check out the docs
> [here](./docs/creating-integrations.md) to learn how.

### Event getters

- Tripit (requires a Tripit Pro subscription)
- Google Calendar
- Webhooks

### Status setters

- Slack

## Getting Started

We're going to use this guide to create a rule that will set your status on
Slack to "🤖 powered by Status Bot" upon creating an event called "Testing
Status Bot" to your Google Calendar (date and time doesn't matter).

### Installing Status

Grab the latest release:

```sh
# Downloads the latest release from this project's GitHub releases page based on
# operating system and processor architecture, then installs it to
# $HOME/.bin and makes it available in the user's current shell
curl -sSL https://raw.githubusercontent.com/carlosonunez/status/HEAD/setup.sh |
    bash
```

### Logging In

Let's log into Slack and Google Calendar. You can either log into everything all
at once:

```sh
status auth login --all
```

Or log into specific integrations:

```sh
status auth login --integrations slack,google_calendar
```

This will start the authentication flow for anything that we're not yet logged
into.

```
➡️  We need to authenticate into "Slack".

Before we begin
================

You'll need to create an app in Slack for Status. To do that, click on this link
and follow the instructions: [link]

Hit ENTER when you've completed the steps shown above: # enter pressed

Let's authenticate!
====================

Enter the client ID for your Slack app: XOXB-blahblahblah
Enter the client secret for your Slack app: blahblahblah
Enter the workspace for your Slack app: some-workspace

Authenticating; please wait...................................................
...............................done!

✅ Logged into Slack!

➡️  We need to authenticate into "Google Calendar".

Before we begin
================

You'll need to create an OAuth client in Google Cloud with the following OAuth
scopes:

- https://www.googleapis.com/auth/calendar.events.readonly
- https://www.googleapis.com/auth/calendar.calendarlist.readonly

Follow the instructions below to do that:

_instructions_

Hit ENTER when you've completed the steps shown above: # enter pressed

Let's authenticate!
====================

Enter the client ID for your Google Calendar OAuth client: blahblah
Enter the client secret for your Google Calendar OAuth client: blahblah

Authenticating; please wait...................................................
...............................done!

Let's wrap up
==============

Select the list of calendars to use for finding events:

[X] Default
[ ] Some calendar
[ ] Some other calendar

✅ Logged into Google Calendar!
```

The authentication will differ between integrations. Run
`status integration show [INTEGRATION_NAME]` to learn more about
authentication works for a specific integration.

### Creating the rule

Now let's create our rule:

```sh
status rule create
```

This will start our new rule wizard:

```
➡️  First, let's give your rule a name.

Enter the name for your new rule: My first rule!

---

➡️  Next, select the sources you would like to retrieve events from:

[x] Google Calendar
[ ] TripIt
[ ] Webhook: Driving
[ ] Webhook: Some other event

---

➡️  Now we're going to configure your event sources.

[Google Calendar] Provide a list of calendar event titles that will trigger this
rule, then hit ENTER to continue. You can also hit ENTER to leave this empty and
trigger on all current-day events:

Testing Status Bot
# enter pressed

[Google Calendar] Enter the MINIMUM length of time an event should be to trigger
this rule. Hit ENTER to skip: # enter pressed

[Google Calendar] Enter the MAXIMUM length of time an event should be to trigger
this rule. Hit ENTER to skip: # enter pressed

[Google Calendar] Should all-day events trigger this rule? ([yes]/no): yes

[Google Calendar] Should this rule ONLY apply to all-day events? (yes/[no]): no

---

➡️  Select the sources you would like to send statuses to:

[x] Slack
[ ] WhatsApp

---

Now we're going to configure transforms that will turn events into statuses.

➡️  Transform 1

[Google Calendar/Slack] Give this transform a name, then hit ENTER: Test transform

[Google Calendar/Slack] Enter text or a regular expression that this
transform will apply to, or hit ENTER to have this apply to all events matched
by this rule: # enter pressed

[Google Calendar/Slack] What status message should this transform generate?
(Regular expression capture groups are supported; DO NOT enter emojis here): powered by
Status bot

[Google Calendar/Slack] Enter an emoji to attach to this status message, or
hit ENTER to skip this: 🤖

[Google Calendar/Slack] Enter a duration that this status should be set for
(e.g. 1h, 60m, 1h30m), or hit ENTER to skip this: # enter pressed

[Google Calendar/Slack] Is this an out-of-office status? (yes/[no]): # enter
pressed

---

Type 'yes' if you want to configure another transform, or hit ENTER to continue (yes/[no]): # enter pressed

---

✅ Rule created!
```

### Running status

We're ready to have Status do its thing!

Run `status start` to run Status in "daemon-mode":

```sh
# Add --log-level=[info|warn|error|debug|trace] to increase logging verbosity.
# Add --log-format=json to show JSON logs. Useful for ingesting into a logging platform.
# Add --interval=[duration] to poll for events more frequently (minimum frequency is every 1m unless configured otherwise by the event getter)
status start
```

This will show:

```
# should display a right-arrow emoji when on the step, a green-checkmark emoji
# if it succeeds, or a red x emoji if it fails with the error shown beneath it

✅ Parsing and testing rules...........looks good!
✅ Checking authentication status for event-getter "Google Calendar".......authenticated!
✅ Checking authentication status for status-setter "Slack".......authenticated!

Status is listening for new events every minute. (Press CTRL-C or send a SIGHUP
signal to stop.)
```

Now, visit Google Calendar and add a calendar event to today called "Testing
Status Bot" of any length.

In about a minute, you'll see something like this:

```

➡️  [2026-03-24T17:12:00.000-07:00] [run_id: OeyRBV4lAphlqVufGHfJL1am] Event from "Google Calendar": {"calendar":"Default","title":"Testing Status Bot","starts_at":1774396800,"ends_at":1774398600,"all_day":false}
➡️  [2026-03-24T17:12:00.500-07:00] [run_id: OeyRBV4lAphlqVufGHfJL1am] Applying transform "Test transform" for status setter "Slack"
➡️  [2026-03-24T17:12:31.000-07:00] [run_id: OeyRBV4lAphlqVufGHfJL1am] Setting status on "Slack": {"status_message":"powered by Status bot","emoji":"🤖","duration":nil,"is_out_of_office":nil}
➡️  [2026-03-24T17:12:31.000-07:00] [run_id: OeyRBV4lAphlqVufGHfJL1am] Setting status on "Slack" was successful, response: {"status":"ok","response":""}
```

Upon going to Slack, you should now see that your status has been set to `🤖 powered by Status bot`. Success!

## Getting Started: A more advanced example

Let's use Status to create a more complex rule that captures all events from Google Calendar and sends different statuses to
Slack and GitHub depending on the events that matched the rule.

> 📝 We'll assume that you've installed Status and authenticated into Google
> Calendar, Slack and GitHub.

### Creating the rule

Since this rule will have several transforms in it, we'll create this rule from YAML to avoid
using the wizard for each condition (and having to re-do the process if we make a mistake!).

```sh
status rule add --use-editor \
  --event-getters=google-calendar \
  --status-setters=slack,github
```

This will open your default editor with a pre-filled template to save time:

```yaml
rule:
  name: A more complex rule
  event_getter:
    name: google_calendar
    params:
        # A list of calendars to use for evaluating this rule. Leave as-is to select events from all calendars.
        calendars: []
        event_names: [] # Calendar events that apply for this rule; leave blank for all events to apply
        min_duration: "" # (optional) The minimum duration a calendar event should be to qualify for this rule
        max_duration: "" # (optional) The maximum duration a calendar event should be to qualify for this rule
        allow_all_day_events: true # include all-day events in this rule.
        only_all_day_events: false # exclude anything that isn't an all-day event from this rule.
  # transforms are processed in the order they are defined!
  transforms:
  - name: In a flight
    pattern: "^([A-Z]{2,3}[0-9]{1,4}) ([A-Z]{3}) to ([A-Z]{3})$" # ex. UAL1 JFK to LAX
    status_setters:
    - name: slack
      params:
        status_message: "In flight: $1: $2 -> $3" # will produce "In flight: UAL1: JFK -> LAX"
        emoji: "✈️ "
  - name: In a meeting
    pattern: ".*"
    status_setter:
      name: slack
      params:
        status_message: In a meeting
        emoji: "📆"
  - name: Test
    pattern: "(.*)"
    status_setter:
      name: dummy
      params:
        status: "My status: $1"
  - name: "Test 2"
    pattern: "(.*)"
    status_setter:
      name: dummy
      params:
        status: "My second status: $1"
```

### Running status

Let's start Status again:

```sh
# Add --log-level=[info|warn|error|debug|trace] to increase logging verbosity.
# Add --log-format=json to show JSON logs. Useful for ingesting into a logging platform.
# Add --interval=[duration] to poll for events more frequently (minimum frequency is every 1m unless configured otherwise by the event getter)
status start
```

This will show:

```
# should display a right-arrow emoji when on the step, a green-checkmark emoji
# if it succeeds, or a red x emoji if it fails with the error shown beneath it

✅ Parsing and testing rules...........looks good!
✅ Checking authentication status for event-getter "Google Calendar".......authenticated!
✅ Checking authentication status for status-setter "Slack".......authenticated!

Status is listening for new events every minute. (Press CTRL-C or send a SIGHUP
signal to stop.)
```

Now, visit Google Calendar and add two overlapping calendar events:

- One called "My meeting" that's two minutes long, and
- Another called "UAL1 JFK to LAX" that's three minutes long.

In about a minute, you'll see something like this if you followed the "Getting Started" guide earlier:

```
➡️  [2026-03-24T17:13:00.000-07:00] [run_id: btP8xQW6HSJ66lTqu5eWQq54] Event from "Google Calendar": {"calendar":"Default","title":"Testing Status Bot","starts_at":1774396800,"ends_at":1774398600,"all_day":false}
➡️  [2026-03-24T17:13:00.000-07:00] [run_id: btP8xQW6HSJ66lTqu5eWQq54] Event from "Google Calendar": {"calendar":"Default","title":"My meeting","starts_at":1774396800,"ends_at":1774398600,"all_day":false}
➡️  [2026-03-24T17:13:00.000-07:00] [run_id: btP8xQW6HSJ66lTqu5eWQq54] Event from "Google Calendar": {"calendar":"Default","title":"UAL1 JFK to LAX","starts_at":1774396800,"ends_at":1774398600,"all_day":false}
➡️  [2026-03-24T17:13:00.500-07:00] [run_id: btP8xQW6HSJ66lTqu5eWQq54] Applying transform "Test transform" for status setter "Slack"
➡️  [2026-03-24T17:13:31.000-07:00] [run_id: btP8xQW6HSJ66lTqu5eWQq54] Setting status on "Slack": {"status_message":"powered by Status bot","emoji":"🤖","duration":nil,"is_out_of_office":nil}
➡️  [2026-03-24T17:13:31.000-07:00] [run_id: btP8xQW6HSJ66lTqu5eWQq54] Setting status on "Slack" was skipped, response: {"status":"ok","response":"status is the same"}
```

This is because every status setter only acts on the FIRST event that's provided
to it during a run. Delete the "Testing Status bot" event from your calendar,
then wait a minute for the next Status run. At that point, you'll see someting like this:

```
➡️  [2026-03-24T17:14:00.000-07:00] [run_id: YjhnLqjU06Kr2ZoNbD6F6BBX] Event from "Google Calendar": {"calendar":"Default","title":"My meeting","starts_at":1774396800,"ends_at":1774398600,"all_day":false}
➡️  [2026-03-24T17:14:00.000-07:00] [run_id: YjhnLqjU06Kr2ZoNbD6F6BBX] Event from "Google Calendar": {"calendar":"Default","title":"UAL1 JFK to LAX","starts_at":1774396800,"ends_at":1774398600,"all_day":false}
➡️  [2026-03-24T17:14:00.250-07:00] [run_id: YjhnLqjU06Kr2ZoNbD6F6BBX] Using "Google Calendar" event "My meeting"
➡️  [2026-03-24T17:14:00.500-07:00] [run_id: YjhnLqjU06Kr2ZoNbD6F6BBX] Not applying transform "In a flight" for status setter "Slack", reason: no matches
➡️  [2026-03-24T17:14:00.500-07:00] [run_id: YjhnLqjU06Kr2ZoNbD6F6BBX] Applying transform "In a meeting" for status setter "Slack"
➡️  [2026-03-24T17:14:31.000-07:00] [run_id: YjhnLqjU06Kr2ZoNbD6F6BBX] Setting status on "Slack": {"status_message":"In a meeting","emoji":"📆","duration":nil,"is_out_of_office":nil}
➡️  [2026-03-24T17:14:31.000-07:00] [run_id: YjhnLqjU06Kr2ZoNbD6F6BBX] Setting status on "Slack" was successful, response: {"status":"ok","response":""}
➡️  [2026-03-24T17:14:31.200-07:00] [run_id: YjhnLqjU06Kr2ZoNbD6F6BBX] Applying transform "Test" for status setter "Dummy"
➡️  [2026-03-24T17:14:31.400-07:00] [run_id: YjhnLqjU06Kr2ZoNbD6F6BBX] Setting status on "Dummy": {"status_message":"My status: My meeting"}
➡️  [2026-03-24T17:14:31.450-07:00] [run_id: YjhnLqjU06Kr2ZoNbD6F6BBX] Setting status on "Dummy" was successful, response: {"status":"ok","response":""}
➡️  [2026-03-24T17:14:31.600-07:00] [run_id: YjhnLqjU06Kr2ZoNbD6F6BBX] Not applying transform "Test 2" for status setter "Dummy", reason: "event already applied"
```

Your status on Slack should be "📆 In a meeting". Wait three minutes, then check again:


```
➡️  [2026-03-24T17:17:00.000-07:00] [run_id: YjhnLqjU06Kr2ZoNbD6F6BBX] Event from "Google Calendar": {"calendar":"Default","title":"UAL1 JFK to LAX","starts_at":1774396800,"ends_at":1774398600,"all_day":false}
➡️  [2026-03-24T17:17:00.250-07:00] [run_id: YjhnLqjU06Kr2ZoNbD6F6BBX] Using "Google Calendar" event "UAL1 JFK to LAX"
➡️  [2026-03-24T17:17:00.500-07:00] [run_id: YjhnLqjU06Kr2ZoNbD6F6BBX] Applying transform "In a flight" for status setter "Slack"
➡️  [2026-03-24T17:17:31.000-07:00] [run_id: YjhnLqjU06Kr2ZoNbD6F6BBX] Setting status on "Slack": {"status_message":"In flight: UAL1: JFK->LAX ","emoji":"✈️ ","duration":nil,"is_out_of_office":nil}
➡️  [2026-03-24T17:17:31.000-07:00] [run_id: YjhnLqjU06Kr2ZoNbD6F6BBX] Setting status on "Slack" was successful, response: {"status":"ok","response":""}
➡️  [2026-03-24T17:17:31.200-07:00] [run_id: YjhnLqjU06Kr2ZoNbD6F6BBX] Applying transform "Test" for status setter "Dummy"
➡️  [2026-03-24T17:17:31.400-07:00] [run_id: YjhnLqjU06Kr2ZoNbD6F6BBX] Setting status on "Dummy": {"status_message":"My status: My meeting"}
➡️  [2026-03-24T17:17:31.450-07:00] [run_id: YjhnLqjU06Kr2ZoNbD6F6BBX] Setting status on "Dummy" was successful, response: {"status":"ok","response":""}
➡️  [2026-03-24T17:17:31.600-07:00] [run_id: YjhnLqjU06Kr2ZoNbD6F6BBX] Not applying transform "Test 2" for status setter "Dummy", reason: "event already applied"
```

Your status should now be "✈️  In flight: UAL1: JFK->LAX".

Pretty cool, right?!?!?!

## How Status Works

### The Event Loop

Status is good at one thing: setting statuses on various platforms
based on events that happen on other platforms. It accomplishes this by doing
the following in a loop until you stop it:

- Reads a list of rules ranked by priority,
- Then, for each rule:
  - Retrieves "events", like meetings, trips and webhooks, from "event getters"
  - Converts data from the first event found from an "event getter" into a status
    message through "transforms"
  - Sends the "transformed" status message to a list of "status setters"
    associated with the rule, and, finally
  - Moves on to the next event IF this rule did not return any events that
    qualify for a status update.

### Components

- An "event getter" is a source from which events are retrieved, like TripIt or
  Google Calendar.

- A "status setter" is a place to which statuses generated by events from "event
  getters" are posted, like Slack, WhatsApp or GitHub.

- "Transforms" are actions that turn data from an event getter into a status
  message that's provided to a "status setter".

  For instance, the example below turns an "Out of Office - Road Trip" all-day
  event on Google Calendar into a "🚗 blasting bebop on max volume" status
  message:

  ```yaml
  transform:
    from:
      source: google_calendar
      from: "Out of Office - Road Trip"
      to: "🚗 blasting bebop on max volume"
  ```

- A "rule" connects event getters to status setters through transforms.

## Using Status

### Browsing integrations

Let's run the command below to see what integrations we have available:

```sh
# Add --format-=json or --format=yaml to see this in JSON or YAML format,
# respectively.
status integration list --all
```

This will yield something like the below:

```
TYPE            NAME                SOURCE          PARAMETERS
event-getter    google_calendar     builtin         - event_name
                                                    - event_date_time
                                                    - min_duration
                                                    - max_duration
                                                    - allow_all_day_events
                                                    - only_all_day_events
status-setter   slack               builtin         - status_message
                                                    - emoji
                                                    - duration
                                                    - set_out_of_office
```

You can also see specific types of integrations, like this:

```sh
# Add --format-=json or --format=yaml to see this in JSON or YAML format,
# respectively.
status integration list --type event-getter
```

Which will show this:


```
TYPE            NAME                SOURCE          PARAMETERS
event-getter    google_calendar     builtin         - event_name
                                                    - event_date_time
                                                    - min_duration
                                                    - max_duration
                                                    - allow_all_day_events
```

Use `status integration show` if you want to learn more about a specific integration:

```sh
# Add --all to the command below to see info about all of the integrations
# Status knows about.
# Add --format=json or --format=yaml to display this as JSON or YAML
# respectively.
status integration show --name slack,google_calendar
```

The output for that will look like this. Add `--paginate` to the end of the
command above to display this in a pager like `less`.

```
NAME: slack
DESCRIPTION: Sets statuses on Slack.
TYPE: event_getter
PARAMETERS

- NAME:        status_message
  DESCRIPTION: The status message to send to Slack.
  TYPE:        string

- NAME:        emoji
  DESCRIPTION: An emoji to attach to the status. Must exist in your workspace.
  TYPE:        string

- NAME:        duration
  DESCRIPTION: How long to set this status for. (Set 'overwrite' to true to
               ignore this.)
  TYPE:        duration

- NAME:        is_out_of_office
  DESCRIPTION: How long to set this status for.
  TYPE:        boolean

- NAME:        overwrite
  DESCRIPTION: Overwrites whatever the current status is.
  TYPE:        boolean

AUTHENTICATION PARAMETERS

- NAME: client_id
  DESCRIPTION: The client ID for your Slack bot or app.
  REQUIRED: yes
  ENVIRONMENT VARIABLE: SLACK_API_CLIENT_ID
- NAME: client_secret
  DESCRIPTION: The client secret for your Slack bot or app.
  REQUIRED: yes
  ENVIRONMENT VARIABLE: SLACK_API_CLIENT_SECRET
- NAME: workspace_name
  DESCRIPTION: The name of the workspace you'll be authenticating into.
  REQUIRED: yes
  ENVIRONMENT VARIABLE: SLACK_WORKSPACE

AUTHENTICATION NOTES

- Your workspace admin will need to approve your Slack app in order for this
  integration to work.

---

NAME: google_calendar
DESCRIPTION: Retrieves events from Google Calendar.
TYPE: status_setter
PARAMETERS

- NAME:        event_name
  DESCRIPTION: The name of the event you want to generate a status from.
               Defaults to all of today's events.
  TYPE:        string

- NAME:        min_duration
  DESCRIPTION: Only select events longer than this duration.
  TYPE:        duration

- NAME:        max_duration
  DESCRIPTION: Only select events shorter than this duration.
  TYPE:        duration

- NAME:        is_out_of_office
  DESCRIPTION: How long to set this status for.
  TYPE:        boolean

- NAME:        allow_all_day_events
  DESCRIPTION: Set this to false if you want to EXCLUDE all day events.
  TYPE:        boolean

- NAME:        only_all_day_events
  DESCRIPTION: Set this to false if you want ONLY all day events.
  TYPE:        boolean

AUTHENTICATION PARAMETERS

- NAME: client_id
  DESCRIPTION: The client ID for your Google Calendar OAuth app.
  REQUIRED: yes
  ENVIRONMENT VARIABLE: GOOGLE_CALENDAR_API_CLIENT_ID
- NAME: client_secret
  DESCRIPTION: The client secret for your Google Calendar OAuth app.
  REQUIRED: yes
  ENVIRONMENT VARIABLE: GOOGLE_CALENDAR_API_CLIENT_SECRET
- NAME: service_account_json_path
  DESCRIPTION: The file path to your Google Cloud service account. Overrides
  `client_id` and `client_secret`.
  REQUIRED: no
  ENVIRONMENT VARIABLE: GOOGLE_CLOUD_SERVICE_ACCOUNT_JSON

AUTHENTICATION NOTES

- You'll need to use the Google Cloud console to generate a Google Calendar
  OAuth app.
```

### Config file

Status stores rules and app configuration into a file called `config.yaml` in the following locations:

- **MacOS** and **Linux**: `$HOME/.config/status/config.yaml`
- **Windows**: `%USERPROFILE%\Roaming\status\config.yaml`

You can edit anything in Status in this file. Run `status check config` to make
sure that your configuration is correct.

## AI Disclaimer

Most of the code powering `status` was written by Claude.

This README was initially written by me. Claude has the ability to modify it based
on what it builds.
