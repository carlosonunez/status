module Status
   module Exceptions
     class ClientNotInitialized < StandardError
       @default_message = %Q(
       The TripIt client has not been initialized. Please do so by using
       Status::Clients::TripIt.new(consumer_token, token_secret).
       )
       def initialize(msg=@default_message)
         super
       end
     end
   end
end
