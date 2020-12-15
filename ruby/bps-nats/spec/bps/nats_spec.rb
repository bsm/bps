require 'bps/nats'
require 'spec_helper'

RSpec.describe 'NATS', nats: true do
  context 'resolve addrs' do
    let(:publisher) { double('BPS::Publisher::NATS') }
    before          { allow(BPS::Publisher::NATS).to receive(:new).and_return(publisher) }

    it 'should resolve simple URLs' do
      expect(BPS::Publisher::NATS)
        .to receive(:new)
        .with(servers: ['nats://test.host:4222'])
        .and_return(publisher)
      BPS::Publisher.resolve(URI.parse('nats://test.host:4222'))
    end

    it 'should resolve URLs with multiple hosts' do
      expect(BPS::Publisher::NATS)
        .to receive(:new)
        .with(servers: ['nats://foo.host:4222', 'nats://bar.host:4222'])
        .and_return(publisher)
      BPS::Publisher.resolve(URI.parse('nats://foo.host,bar.host:4222'))
    end

    it 'should resolve URLs with multiple hosts/ports' do
      expect(BPS::Publisher::NATS)
        .to receive(:new)
        .with(servers: ['nats://foo.host:4223', 'nats://bar.host:4222'])
        .and_return(publisher)
      BPS::Publisher.resolve(URI.parse('nats://foo.host%3A4223,bar.host'))
    end
  end

  context BPS::Publisher::NATS do
    let(:nats_servers) { ENV.fetch('NATS_SERVERS', '127.0.0.1:4222').split(',') }
    let(:nats_servers_with_scheme) { nats_servers.map {|s| "nats://#{s}" } }

    let(:publisher_url) { "nats://#{CGI.escape(nats_servers.join(','))}" }

    # Pure NATS doesn't retain messages, so we need to subscribe BEFORE emitting.
    # Also, subscription is asynchronous, so we have to use blocking queue to wait for messages.

    # connect to NATS before any test code runs:
    let!(:nats_client) do
      opts = {
        servers: nats_servers_with_scheme,
        dont_randomize_servers: true,
      }
      client = ::NATS::IO::Client.new
      client.connect(opts)
      client
    end

    # don't forget to close NATS connection after tests are done:
    after do
      nats_client.close
    end

    # blocking queue to gather messages in background thread:
    let(:messages_queue) { Queue.new }

    # subscribe to test topic, non-blocking; messages will be pushed into blocking queue:
    def prepare_topic(topic_name, num_messages)
      nats_client.subscribe(topic_name, max: num_messages) {|msg| messages_queue << msg }
    end

    # simply drain messages queue to get messages:
    def read_messages(_topic_name, num_messages)
      num_messages.times.map { messages_queue.pop }
    end

    it_behaves_like 'publisher'
  end
end
