$LOAD_PATH << '../lib'
require 'rspec'
require 'dotenv'
require 'yaml'
require 'status'

# Retrieves a test fixture from the spec fixtures path.
# This is used so that individual tests don't have to hardcode the spec file each time.
def get_fixture_file_path(path)
  return "spec/fixtures/#{path}"
end

raise 'Test .env not found' if !File.exist? '.env.test'
Dotenv.load('.env.test')

RSpec.configure do |configuration|
end
