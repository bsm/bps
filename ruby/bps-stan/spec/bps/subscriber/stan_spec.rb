require 'bps/stan'
require 'spec_helper'

RSpec.describe 'STAN', stan: true do
  context 'with addr resolving' do
    let(:subscriber) { instance_double('BPS::Subscriber::STAN') }

    before { allow(BPS::Subscriber::STAN).to receive(:new).and_return(subscriber) }

    it 'resolves simple URLs' do
      allow(BPS::Subscriber::STAN)
        .to receive(:new)
        .with('CLUSTER', 'CLIENT', nats: { servers: ['nats://test.host:4222'] })
        .and_return(subscriber)
      expect(BPS::Subscriber.resolve(URI.parse('stan://test.host:4222?cluster_id=CLUSTER&client_id=CLIENT')))
        .to eq(subscriber)
    end

    it 'resolves URLs with multiple hosts' do
      allow(BPS::Subscriber::STAN)
        .to receive(:new)
        .with('CLUSTER', 'CLIENT', nats: { servers: ['nats://foo.host:4222', 'nats://bar.host:4222'] })
        .and_return(subscriber)
      expect(BPS::Subscriber.resolve(URI.parse('stan://foo.host,bar.host:4222?cluster_id=CLUSTER&client_id=CLIENT')))
        .to eq(subscriber)
    end

    it 'resolves URLs with multiple hosts/ports' do
      allow(BPS::Subscriber::STAN)
        .to receive(:new)
        .with('CLUSTER', 'CLIENT', nats: { servers: ['nats://foo.host:4223', 'nats://bar.host:4222'] })
        .and_return(subscriber)
      expect(BPS::Subscriber.resolve(URI.parse('stan://foo.host%3A4223,bar.host?cluster_id=CLUSTER&client_id=CLIENT')))
        .to eq(subscriber)
    end
  end

  context BPS::Subscriber::STAN do
    let(:cluster_id) { 'test-cluster' } # this is a default cluster for https://hub.docker.com/_/nats-streaming
    let(:client_id)  { 'bps-test' }

    let(:nats_servers) { ENV.fetch('STAN_SERVERS', '127.0.0.1:4222').split(',') }
    let(:nats_servers_with_scheme) { nats_servers.map {|s| "nats://#{s}" } }

    let(:subscriber_url) { "stan://#{CGI.escape(nats_servers.join(','))}/?cluster_id=#{cluster_id}&client_id=#{client_id}" }

    def produce_messages(topic_name, messages)
      opts = {
        servers: nats_servers_with_scheme,
        dont_randomize_servers: true,
      }
      ::STAN::Client.new.connect(cluster_id, client_id, nats: opts) do |client|
        messages.each do |message|
          client.publish(topic_name, message)
        end
      end
    end

    it_behaves_like 'subscriber'
  end
end
