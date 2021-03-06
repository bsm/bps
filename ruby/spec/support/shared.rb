require 'securerandom'

RSpec.shared_examples 'publisher' do
  # WARNING: This example group requires the following helpers to be defined by caller:
  #   - `read_messages(topic_name, num_messages)`
  #   - `publisher_url`

  subject do
    BPS::Publisher.resolve(publisher_url)
  end

  after do
    subject.close
  end

  it 'registers' do
    expect(subject).to be_a(described_class)
  end

  it 'publishes' do
    topic_name = "bps-test-topic-#{SecureRandom.uuid}"
    messages = Array.new(3) { "bps-test-message-#{SecureRandom.uuid}" }

    # call optional `setup_topic` - it's needed only adapters
    # that don't retain messages so subscribing must be done before publishing:
    begin
      setup_topic(topic_name, messages.count)
    rescue NoMethodError # rubocop:disable Lint/SuppressedException
    end

    topic = subject.topic(topic_name)
    messages.each {|msg| topic.publish(msg) }
    subject.close

    published = read_messages(topic_name, messages.count)
    expect(published).to match_array(messages)
  end
end

RSpec.shared_examples 'subscriber' do
  # WARNING: This example group requires the following helpers to be defined by caller:
  #   - `produce_messages(topic_name, messages)` (messages - Array[String])
  #   - `subscriber_url`

  subject do
    BPS::Subscriber.resolve(subscriber_url)
  end

  after do
    subject.close
  end

  it 'registers' do
    expect(subject).to be_a(described_class)
  end

  it 'publishes' do
    topic_name = "bps-test-topic-#{SecureRandom.uuid}"
    messages = Array.new(3) { "bps-test-message-#{SecureRandom.uuid}" }

    produce_messages(topic_name, messages)

    # TODO: ideally, need to pre-define/convert canonical subscription options
    consumed = Queue.new
    subject.subscribe(topic_name, start_at: :first) do |msg_data|
      consumed.push(msg_data)
    end

    consumed_messages = Array.new(3) { consumed.pop }
    expect(consumed_messages).to match_array(messages)
  end
end
