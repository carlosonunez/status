require 'tripit'
require 'status/exceptions'

module Status
  module Clients
    class TripIt
      @client = nil

      def authenticate(user_id: nil,
                       consumer_token: nil,
                       token_secret: nil)

      end

      def initialize(consumer_token: nil, token_secret: nil)
        return @client if self.initialized?

        begin
          @client = TripIt::OAuth.new(consumer_token, token_secret)
          client.authorize_from_access(access_token, access_token_secret)
          return @client
        rescue
          raise "Failed to initialize the TripIt client."
        end
      end

      def get_trips()
        # pending test
      end

      def initialized?
        begin
          _ = TripIt::Profile.new(self).screen_name
          return true
        rescue
          return false
        end
      end
    end
  end
end
