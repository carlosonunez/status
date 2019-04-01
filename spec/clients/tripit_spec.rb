require 'spec_helper'

describe 'Given a TripIt client' do
  before(:all) do
    @test_tripit_id = "test_user_1"
    @expected_tripit_trips = YAML.load_file(get_fixture_file_path 'tripit/trips.yaml')
  end

  context 'When a user tries to authorize Status for the first time' do
    it 'Should persist their access token and token secret' do
      ( access_token, access_secret ) =
        Status::Clients::TripIt.authenticate(user_id: @test_tripit_id,
                                             consumer_token: ENV['TRIPIT_CONSUMER_TOKEN'],
                                             token_secret: ENV['TRIPIT_TOKEN_SECRET'])
      expect(access_token).to eq(@expected_access_token)
      expect(access_secret).to eq(@expected_access_secret)
    end
  end

  context 'When a user has already authorized Status' do
    it 'Should be able to enumerate all of my trips' do
      pending("Unfortunately, we have to take care of OAuth first.")
      @tripit_client = Status::Clients::TripIt.new(
        user_id: @test_tripit_id,
        consumer_token: ENV['TRIPIT_CONSUMER_TOKEN'],
        token_secret: ENV['TRIPIT_TOKEN_SECRET']
      )
      @trips_found = @tripit_client.get_trips
      expect(@trips_found).to eq(@expected_tripit_trips)
    end
  end
end
