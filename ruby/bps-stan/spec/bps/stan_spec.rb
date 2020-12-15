require 'bps/stan'
require 'spec_helper'

RSpec.describe 'STAN', stan: true do
  context 'resolve addrs' do
    let(:publisher) { double('BPS::Publisher::STAN') }
    before          { allow(BPS::Publisher::STAN).to receive(:new).and_return(publisher) }

    it 'should resolve simple URLs' do
      expect(BPS::Publisher::STAN)
        .to receive(:new)
        .with('CLUSTER', 'CLIENT', nats: { servers: ['nats://test.host:4222'] })
        .and_return(publisher)
      BPS::Publisher.resolve(URI.parse('stan://test.host:4222?cluster_id=CLUSTER&client_id=CLIENT'))
    end

    it 'should resolve URLs with multiple hosts' do
      expect(BPS::Publisher::STAN)
        .to receive(:new)
        .with('CLUSTER', 'CLIENT', nats: { servers: ['nats://foo.host:4222', 'nats://bar.host:4222'] })
        .and_return(publisher)
      BPS::Publisher.resolve(URI.parse('stan://foo.host,bar.host:4222?cluster_id=CLUSTER&client_id=CLIENT'))
    end

    it 'should resolve URLs with multiple hosts/ports' do
      expect(BPS::Publisher::STAN)
        .to receive(:new)
        .with('CLUSTER', 'CLIENT', nats: { servers: ['nats://foo.host:4223', 'nats://bar.host:4222'] })
        .and_return(publisher)
      BPS::Publisher.resolve(URI.parse('stan://foo.host%3A4223,bar.host?cluster_id=CLUSTER&client_id=CLIENT'))
    end
  end

  context BPS::Publisher::STAN do
    let(:cluster_id) { 'test-cluster' } # this is a default cluster for https://hub.docker.com/_/nats-streaming
    let(:client_id)  { 'bps-test' }

    let(:nats_servers) { ENV.fetch('STAN_SERVERS', '127.0.0.1:4222').split(',') }
    let(:nats_servers_with_scheme) { nats_servers.map {|s| "nats://#{s}" } }

    let(:publisher_url) { "stan://#{CGI.escape(nats_servers.join(','))}/?cluster_id=#{cluster_id}&client_id=#{client_id}" }

    def read_messages(topic_name, num_messages)
      [].tap do |messages|
        opts = {
          servers: nats_servers_with_scheme,
          dont_randomize_servers: true,
        }
        ::STAN::Client.new.connect(cluster_id, client_id, nats: opts) do |client|
          client.subscribe(topic_name, start_at: :first) do |msg|
            messages << msg.data
            next if messages.size == num_messages
          end
        end
      end
    end

    it_behaves_like 'publisher'
  end
end
