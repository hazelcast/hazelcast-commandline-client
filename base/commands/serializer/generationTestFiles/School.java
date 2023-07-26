package com.education;

import com.hazelcast.nio.serialization.compact.CompactReader;
import com.hazelcast.nio.serialization.compact.CompactSerializer;
import com.hazelcast.nio.serialization.compact.CompactWriter;

import com.rooms.Classroom;

import javax.annotation.Nonnull;
import java.util.Arrays;
import java.util.Objects;

public class School {

    public static final class Serializer implements CompactSerializer<School> {
        @Nonnull
        @Override
        public School read(@Nonnull CompactReader reader) {
            int id = reader.readInt32("id");
            com.rooms.Classroom[] classrooms = reader.readArrayOfCompact("classrooms", com.rooms.Classroom.class);
            return new School(id, classrooms);
        }

        @Override
        public void write(@Nonnull CompactWriter writer, @Nonnull School object) {
            writer.writeInt32("id", object.id);
            writer.writeArrayOfCompact("classrooms", object.classrooms);
        }
    };

    public static final CompactSerializer<School> HZ_COMPACT_SERIALIZER = new Serializer();

    private int id;
    private com.rooms.Classroom[] classrooms;

    public School() {
    }

    public School(int id, com.rooms.Classroom[] classrooms) {
        this.id = id;
        this.classrooms = classrooms;
    }

    public int getId() {
        return id;
    }

    public com.rooms.Classroom[] getClassrooms() {
        return classrooms;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;

        School that = (School) o;
        if (id != that.id) return false;
        if (!Arrays.equals(classrooms, that.classrooms)) return false;

        return true;
    }

    @Override
    public int hashCode() {
        int result = 0;
        result = 31 * result + (int) id;
        result = 31 * result + Arrays.hashCode(classrooms);

        return result;
    }

    @Override
    public String toString() {
        return "<School> {"
                + "id=" + id
                + ", classrooms=" + Arrays.toString(classrooms)
                + '}';
    }

}