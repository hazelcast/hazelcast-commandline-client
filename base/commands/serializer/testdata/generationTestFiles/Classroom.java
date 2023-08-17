package com.rooms;

import com.hazelcast.nio.serialization.compact.CompactReader;
import com.hazelcast.nio.serialization.compact.CompactSerializer;
import com.hazelcast.nio.serialization.compact.CompactWriter;

import com.entities.Student;

import javax.annotation.Nonnull;
import java.util.Arrays;
import java.util.Objects;

public class Classroom {

    public static final class Serializer implements CompactSerializer<Classroom> {
        @Nonnull
        @Override
        public Classroom read(@Nonnull CompactReader reader) {
            int id = reader.readInt32("id");
            com.entities.Student[] students = reader.readArrayOfCompact("students", com.entities.Student.class);
            return new Classroom(id, students);
        }

        @Override
        public void write(@Nonnull CompactWriter writer, @Nonnull Classroom object) {
            writer.writeInt32("id", object.id);
            writer.writeArrayOfCompact("students", object.students);
        }
    };

    public static final CompactSerializer<Classroom> HZ_COMPACT_SERIALIZER = new Serializer();

    private int id;
    private com.entities.Student[] students;

    public Classroom() {
    }

    public Classroom(int id, com.entities.Student[] students) {
        this.id = id;
        this.students = students;
    }

    public int getId() {
        return id;
    }

    public com.entities.Student[] getStudents() {
        return students;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;

        Classroom that = (Classroom) o;
        if (id != that.id) return false;
        if (!Arrays.equals(students, that.students)) return false;

        return true;
    }

    @Override
    public int hashCode() {
        int result = 0;
        result = 31 * result + (int) id;
        result = 31 * result + Arrays.hashCode(students);

        return result;
    }

    @Override
    public String toString() {
        return "<Classroom> {"
                + "id=" + id
                + ", students=" + Arrays.toString(students)
                + '}';
    }

}