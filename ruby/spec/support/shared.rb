require 'securerandom'

RSpec.shared_examples 'publisher' do |features|
  # read_messages must read message data (as array of strings/byte-slices) from given `_topic_name`.
  # `_num_messages` tells how many messages were produced to given `_topic_name`.
  def read_messages(_topic_name, _num_messages)
    raise 'must be overridden'
  end

  before do
    expect(features).to have_key(:url) # each publisher must register under at least 1 scheme
  end

  subject do
    BPS::Publisher.resolve(features[:url])
  end

  after do
    subject.close
  end

  it 'should register' do
    expect(subject).to be_a(described_class)
  end

  it 'should publish' do
    topic_name = "bps-test-topic-#{SecureRandom.uuid}"
    messages = 3.times.map { "bps-test-message-#{SecureRandom.uuid}" }

    topic = subject.topic(topic_name)
    messages.each {|msg| topic.publish(msg) }
    subject.close

    published = read_messages(topic_name, messages.count)
    expect(published).to match_array(messages)
  end
end
