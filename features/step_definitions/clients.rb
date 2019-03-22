Given("a Slack client") do
  pending # Write code here that turns the phrase above into concrete actions
end

Given("a TripIt trip called {string}") do |trip_name|
  @tripit_client = Status::Clients::TripIt.new()
  @matching_tripit_trip = @tripit_client.get_trip(trip_name)
end

