package com.example;

import com.hazelcast.nio.serialization.compact.CompactReader;
import com.hazelcast.nio.serialization.compact.CompactSerializer;
import com.hazelcast.nio.serialization.compact.CompactWriter;

import javax.annotation.Nonnull;
import java.util.Arrays;
import java.util.Objects;

public class Example1 {

    public static final class Serializer implements CompactSerializer<Example1> {
        @Nonnull
        @Override
        public Example1 read(@Nonnull CompactReader reader) {
            Example2 example = reader.readCompact("example");
            com.example.Example3[] examples = reader.readArrayOfCompact("examples", com.example.Example3.class);
            return new Example1(example, examples);
        }

        @Override
        public void write(@Nonnull CompactWriter writer, @Nonnull Example1 object) {
            writer.writeCompact("example", object.example);
            writer.writeArrayOfCompact("examples", object.examples);
        }
    };

    public static final CompactSerializer<Example1> HZ_COMPACT_SERIALIZER = new Serializer();

    private Example2 example;
    private com.example.Example3[] examples;

    public Example1() {
    }

    public Example1(Example2 example, com.example.Example3[] examples) {
        this.example = example;
        this.examples = examples;
    }

    public Example2 getExample() {
        return example;
    }

    public com.example.Example3[] getExamples() {
        return examples;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;

        Example1 that = (Example1) o;
        if (!Objects.equals(example, that.example)) return false;
        if (!Arrays.equals(examples, that.examples)) return false;

        return true;
    }

    @Override
    public int hashCode() {
        int result = 0;
        result = 31 * result + Objects.hashCode(example);
        result = 31 * result + Arrays.hashCode(examples);

        return result;
    }

    @Override
    public String toString() {
        return "<Example1> {"
                + "example=" + example
                + ", examples=" + Arrays.toString(examples)
                + '}';
    }

}