@wip
Feature: Signing up for Status

  We want getting started with Status to be a seamless experience through Slack.
  Eventually I would like to use OAuth for a more secure sign-on process.

  Scenario: Trying to sign up through Slackbot
    Given a Slack client
    When I start a chat with "#slackbot"
    And I type "/givemestatus"
    Then I am asked for my TripIt API key
    And I am asked for an initial filter
    And my status should be updated within 30 seconds.

  Scenario: Trying to sign up from any other channel
    Given a Slack client
    When I start a chat with "#not-slackbot"
    And I type "/givemestatus"
    Then I should see a private message saying "To sign up for Status, DM '/givemestatus' to Slackbot."
