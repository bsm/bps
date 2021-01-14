require 'spec_helper'

RSpec.describe BPS::Coercer do
  subject { coercer }

  let(:coercer) do
    described_class.new(
      name: :string,
      codec: :symbol,
      retries: :int,
      backoff: :float,
      idempotent: :bool,
      tags: [:string],
    )
  end
  let :options do
    {
      name: 123,
      codec: 'snappy',
      retries: '4',
      backoff: '10',
      idempotent: '1',
      tags: [:foo, 33],
      extra: 'foo',
    }
  end

  it 'validates' do
    expect { described_class.new(name: :unknown) }.to raise_error(ArgumentError, /Unknown type :unknown/)
    expect { described_class.new(bad: []) }.to raise_error(ArgumentError, /Array types must have exactly one entry/)
  end

  it 'coerces options' do
    expect(coercer.coerce(options)).to eq(
      name: '123',
      codec: :snappy,
      retries: 4,
      backoff: 10.0,
      idempotent: true,
      tags: %w[foo 33],
    )
  end
end
