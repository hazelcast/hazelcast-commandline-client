package com.example;

import com.hazelcast.nio.serialization.compact.CompactReader;
import com.hazelcast.nio.serialization.compact.CompactSerializer;
import com.hazelcast.nio.serialization.compact.CompactWriter;

import javax.annotation.Nonnull;
import java.util.Arrays;
import java.util.Objects;

public class Example3 {

    public static final class Serializer implements CompactSerializer<Example3> {
        @Nonnull
        @Override
        public Example3 read(@Nonnull CompactReader reader) {
            long bar = reader.readInt64("bar");
            return new Example3(bar);
        }

        @Override
        public void write(@Nonnull CompactWriter writer, @Nonnull Example3 object) {
            writer.writeInt64("bar", object.bar);
        }
    };

    public static final CompactSerializer<Example3> HZ_COMPACT_SERIALIZER = new Serializer();

    private long bar;

    public Example3() {
    }

    public Example3(long bar) {
        this.bar = bar;
    }

    public long getBar() {
        return bar;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;

        Example3 that = (Example3) o;
        if (bar != that.bar) return false;

        return true;
    }

    @Override
    public int hashCode() {
        int result = 0;
        result = 31 * result + (int) (bar ^ (bar >>> 32));

        return result;
    }

    @Override
    public String toString() {
        return "<Example3> {"
                + "bar=" + bar
                + '}';
    }

}