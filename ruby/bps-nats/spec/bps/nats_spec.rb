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

    def read_messages(topic_name, num_messages)
      opts = {
        servers: nats_servers_with_scheme,
        dont_randomize_servers: true,
      }
      client = ::NATS::IO::Client.new
      client.connect(opts)

      # subscriptions are asynchronous, so using blocking queue to wait for messages:
      messages = Queue.new
      client.subscribe(topic_name, max: num_messages) { |msg| messages << msg }
      num_messages.times.map{ messages.pop }
    ensure
      client&.close
      messages&.close
    end

    it_behaves_like 'publisher'
  end
end
