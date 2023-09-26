package com.entities;

import com.hazelcast.nio.serialization.compact.CompactReader;
import com.hazelcast.nio.serialization.compact.CompactSerializer;
import com.hazelcast.nio.serialization.compact.CompactWriter;

import javax.annotation.Nonnull;
import java.util.Arrays;
import java.util.Objects;

public class Student {

    public static final class Serializer implements CompactSerializer<Student> {
        @Nonnull
        @Override
        public Student read(@Nonnull CompactReader reader) {
            java.lang.String name = reader.readString("name");
            short number = reader.readInt16("number");
            return new Student(name, number);
        }

        @Override
        public void write(@Nonnull CompactWriter writer, @Nonnull Student object) {
            writer.writeString("name", object.name);
            writer.writeInt16("number", object.number);
        }

        @Override
        public Class<Student> getCompactClass() {
            return Student.class;
        }

        @Override
        public String getTypeName() {
            return "00000000-0000-000a-0000-00000000000a";
        }
    };

    public static final CompactSerializer<Student> HZ_COMPACT_SERIALIZER = new Serializer();

    private java.lang.String name;
    private short number;

    public Student() {
    }

    public Student(java.lang.String name, short number) {
        this.name = name;
        this.number = number;
    }

    public java.lang.String getName() {
        return name;
    }

    public short getNumber() {
        return number;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;

        Student that = (Student) o;
        if (!Objects.equals(name, that.name)) return false;
        if (number != that.number) return false;

        return true;
    }

    @Override
    public int hashCode() {
        int result = 0;
        result = 31 * result + Objects.hashCode(name);
        result = 31 * result + (int) number;

        return result;
    }

    @Override
    public String toString() {
        return "<Student> {"
                + "name=" + name
                + ", number=" + number
                + '}';
    }

}