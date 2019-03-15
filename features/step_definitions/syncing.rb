require 'status/clients/tripit'

Given("a TripIt trip called {string}") do |trip_name|
  @tripit_client = Status::Clients::TripIt.new()
  @matching_tripit_trip = @tripit_client.get_trip(trip_name)
end

Given("a Slack client") do
  pending
end

When("I start a chat with {string}") do |string|
  pending # Write code here that turns the phrase above into concrete actions
end

When("I type {string}") do |string|
  pending # Write code here that turns the phrase above into concrete actions
end

Then("I am asked for my TripIt API key") do
  pending # Write code here that turns the phrase above into concrete actions
end

Then("I am asked for an initial filter") do
  pending # Write code here that turns the phrase above into concrete actions
end

Then("my status should be updated within {int} seconds.") do |int|
  pending # Write code here that turns the phrase above into concrete actions
end

Then("I should see a private message saying {string}") do |string|
  pending # Write code here that turns the phrase above into concrete actions
end

When("Status runs") do
  pending # Write code here that turns the phrase above into concrete actions
end

When("it runs during the dates for that trip") do
  pending # Write code here that turns the phrase above into concrete actions
end

When("I am flying") do
  pending # Write code here that turns the phrase above into concrete actions
end

Then("my Slack status should say: {string}") do |string|
  pending # Write code here that turns the phrase above into concrete actions
end

When("I am not flying") do
  pending # Write code here that turns the phrase above into concrete actions
end

Then("my Slack status should say {string}") do |string|
  pending # Write code here that turns the phrase above into concrete actions
end

When("it runs outside of the dates for that trip") do
  pending # Write code here that turns the phrase above into concrete actions
end

Then("my Slack status sholud say {string}") do |string|
  pending # Write code here that turns the phrase above into concrete actions
end

