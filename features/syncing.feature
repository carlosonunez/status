Feature: Syncing status with TripIt and SLack
  Scenario: People should know when I'm on a flight.
    Given a TripIt trip called "Company: Client"
    When Status runs
    And it runs during the dates for that trip
    And I am flying
    Then my Slack status should say: "[Client] :plane: DFW -> LGA"

  @wip
  Scenario: People should know when I'm at a client when I'm not flying.
    Given a TripIt trip called "Company: Client"
    When Status runs
    And it runs during the dates for that trip
    And I am not flying
    Then my Slack status should say "[Bar] 3/1 - 3/4"

  @wip
  Scenario: People should know when I'm home.
    Given a TripIt trip called "Company: Client"
    When Status runs
    And it runs outside of the dates for that trip
    Then my Slack status sholud say ":home:"
