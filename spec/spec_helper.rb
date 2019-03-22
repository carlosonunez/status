require 'rspec'
require 'status'
require 'dotenv'

raise 'Test .env not found' if !File.exist? '.env.test'
Dotenv.load('.env.test')

RSpec.configure do |configuration|
end
