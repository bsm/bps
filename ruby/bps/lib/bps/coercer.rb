require 'set'

module BPS
  class Coercer
    TRUE_VALUES = [true, 'TRUE', 'true', 'T', 't', 1, '1'].to_set

    attr_reader :schema

    def initialize(schema)
      validate!(schema)
      @schema = schema
    end

    def coerce(hash)
      coerce_with(@schema, hash)
    end

    private

    def validate!(schema)
      schema.each do |key, type|
        case type
        when :string, :symbol, :int, :float, :bool
          # OK
        when Hash
          validate!(type)
        when Array
          raise ArgumentError, "Array types must have exactly one entry, but was (#{key} => #{type.inspect})" unless type.size == 1
        else
          raise ArgumentError, "Unknown type #{type.inspect}"
        end
      end
    end

    def coerce_with(schema, hash)
      clone = {}
      schema.each do |key, type|
        next unless hash.key?(key)

        clone[key] = coerce_value(type, hash[key])
      end
      clone
    end

    def coerce_value(type, val)
      case type
      when :string
        val&.to_s
      when :symbol
        val&.to_sym
      when :int
        val&.to_i
      when :float
        val&.to_f
      when :bool
        val.nil? ? nil : TRUE_VALUES.include?(val)
      when Hash
        val.is_a?(Hash) ? coerce_with(type, val) : nil
      when Array
        Array(val).map {|v| coerce_value(type[0], v) }
      end
    end
  end
end
