package com.example;

import com.hazelcast.nio.serialization.compact.CompactReader;
import com.hazelcast.nio.serialization.compact.CompactSerializer;
import com.hazelcast.nio.serialization.compact.CompactWriter;

import javax.annotation.Nonnull;
import java.util.Arrays;
import java.util.Objects;

public class AllTypes {

    public static final class Serializer implements CompactSerializer<AllTypes> {
        @Nonnull
        @Override
        public AllTypes read(@Nonnull CompactReader reader) {
            boolean mboolean = reader.readBoolean("mboolean");
            boolean[] mbooleanArray = reader.readArrayOfBoolean("mbooleanArray");
            byte mbyte = reader.readInt8("mbyte");
            byte[] mbyteArray = reader.readArrayOfInt8("mbyteArray");
            short mshort = reader.readInt16("mshort");
            short[] mshortArray = reader.readArrayOfInt16("mshortArray");
            int mint = reader.readInt32("mint");
            int[] mintArray = reader.readArrayOfInt32("mintArray");
            long mlong = reader.readInt64("mlong");
            long[] mlongArray = reader.readArrayOfInt64("mlongArray");
            float mfloat = reader.readFloat32("mfloat");
            float[] mfloatArray = reader.readArrayOfFloat32("mfloatArray");
            double mdouble = reader.readFloat64("mdouble");
            double[] mdoubleArray = reader.readArrayOfFloat64("mdoubleArray");
            java.lang.String mstring = reader.readString("mstring");
            java.lang.String[] mstringArray = reader.readArrayOfString("mstringArray");
            java.time.LocalDate mdate = reader.readDate("mdate");
            java.time.LocalDate[] mdateArray = reader.readArrayOfDate("mdateArray");
            java.time.LocalTime mtime = reader.readTime("mtime");
            java.time.LocalTime[] mtimeArray = reader.readArrayOfTime("mtimeArray");
            java.time.LocalDateTime mtimestamp = reader.readTimestamp("mtimestamp");
            java.time.LocalDateTime[] mtimestampArray = reader.readArrayOfTimestamp("mtimestampArray");
            java.time.OffsetDateTime mtimestampWithTimezone = reader.readTimestampWithTimezone("mtimestampWithTimezone");
            java.time.OffsetDateTime[] mtimestampWithTimezoneArray = reader.readArrayOfTimestampWithTimezone("mtimestampWithTimezoneArray");
            Boolean mnullableboolean = reader.readNullableBoolean("mnullableboolean");
            Boolean[] mnullablebooleanArray = reader.readArrayOfNullableBoolean("mnullablebooleanArray");
            Byte mnullablebyte = reader.readNullableInt8("mnullablebyte");
            Byte[] mnullablebyteArray = reader.readArrayOfNullableInt8("mnullablebyteArray");
            Short mnullableshort = reader.readNullableInt16("mnullableshort");
            Short[] mnullableshortArray = reader.readArrayOfNullableInt16("mnullableshortArray");
            Integer mnullableint = reader.readNullableInt32("mnullableint");
            Integer[] mnullableintArray = reader.readArrayOfNullableInt32("mnullableintArray");
            Long mnullablelong = reader.readNullableInt64("mnullablelong");
            Long[] mnullablelongArray = reader.readArrayOfNullableInt64("mnullablelongArray");
            Float mnullablefloat = reader.readNullableFloat32("mnullablefloat");
            Float[] mnullablefloatArray = reader.readArrayOfNullableFloat32("mnullablefloatArray");
            Double mnullabledouble = reader.readNullableFloat64("mnullabledouble");
            Double[] mnullabledoubleArray = reader.readArrayOfNullableFloat64("mnullabledoubleArray");
            NestedCompact mcompact = reader.readCompact("mcompact");
            NestedCompact[] mcompactArray = reader.readArrayOfCompact("mcompactArray", NestedCompact.class);
            return new AllTypes(mboolean, mbooleanArray, mbyte, mbyteArray, mshort, mshortArray, mint, mintArray, mlong, mlongArray, mfloat, mfloatArray, mdouble, mdoubleArray, mstring, mstringArray, mdate, mdateArray, mtime, mtimeArray, mtimestamp, mtimestampArray, mtimestampWithTimezone, mtimestampWithTimezoneArray, mnullableboolean, mnullablebooleanArray, mnullablebyte, mnullablebyteArray, mnullableshort, mnullableshortArray, mnullableint, mnullableintArray, mnullablelong, mnullablelongArray, mnullablefloat, mnullablefloatArray, mnullabledouble, mnullabledoubleArray, mcompact, mcompactArray);
        }

        @Override
        public void write(@Nonnull CompactWriter writer, @Nonnull AllTypes object) {
            writer.writeBoolean("mboolean", object.mboolean);
            writer.writeArrayOfBoolean("mbooleanArray", object.mbooleanArray);
            writer.writeInt8("mbyte", object.mbyte);
            writer.writeArrayOfInt8("mbyteArray", object.mbyteArray);
            writer.writeInt16("mshort", object.mshort);
            writer.writeArrayOfInt16("mshortArray", object.mshortArray);
            writer.writeInt32("mint", object.mint);
            writer.writeArrayOfInt32("mintArray", object.mintArray);
            writer.writeInt64("mlong", object.mlong);
            writer.writeArrayOfInt64("mlongArray", object.mlongArray);
            writer.writeFloat32("mfloat", object.mfloat);
            writer.writeArrayOfFloat32("mfloatArray", object.mfloatArray);
            writer.writeFloat64("mdouble", object.mdouble);
            writer.writeArrayOfFloat64("mdoubleArray", object.mdoubleArray);
            writer.writeString("mstring", object.mstring);
            writer.writeArrayOfString("mstringArray", object.mstringArray);
            writer.writeDate("mdate", object.mdate);
            writer.writeArrayOfDate("mdateArray", object.mdateArray);
            writer.writeTime("mtime", object.mtime);
            writer.writeArrayOfTime("mtimeArray", object.mtimeArray);
            writer.writeTimestamp("mtimestamp", object.mtimestamp);
            writer.writeArrayOfTimestamp("mtimestampArray", object.mtimestampArray);
            writer.writeTimestampWithTimezone("mtimestampWithTimezone", object.mtimestampWithTimezone);
            writer.writeArrayOfTimestampWithTimezone("mtimestampWithTimezoneArray", object.mtimestampWithTimezoneArray);
            writer.writeNullableBoolean("mnullableboolean", object.mnullableboolean);
            writer.writeArrayOfNullableBoolean("mnullablebooleanArray", object.mnullablebooleanArray);
            writer.writeNullableInt8("mnullablebyte", object.mnullablebyte);
            writer.writeArrayOfNullableInt8("mnullablebyteArray", object.mnullablebyteArray);
            writer.writeNullableInt16("mnullableshort", object.mnullableshort);
            writer.writeArrayOfNullableInt16("mnullableshortArray", object.mnullableshortArray);
            writer.writeNullableInt32("mnullableint", object.mnullableint);
            writer.writeArrayOfNullableInt32("mnullableintArray", object.mnullableintArray);
            writer.writeNullableInt64("mnullablelong", object.mnullablelong);
            writer.writeArrayOfNullableInt64("mnullablelongArray", object.mnullablelongArray);
            writer.writeNullableFloat32("mnullablefloat", object.mnullablefloat);
            writer.writeArrayOfNullableFloat32("mnullablefloatArray", object.mnullablefloatArray);
            writer.writeNullableFloat64("mnullabledouble", object.mnullabledouble);
            writer.writeArrayOfNullableFloat64("mnullabledoubleArray", object.mnullabledoubleArray);
            writer.writeCompact("mcompact", object.mcompact);
            writer.writeArrayOfCompact("mcompactArray", object.mcompactArray);
        }

        @Override
        public Class<AllTypes> getCompactClass() {
            return AllTypes.class;
        }

        @Override
        public String getTypeName() {
            return "00000000-0000-000a-0000-00000000000a";
        }
    };

    public static final CompactSerializer<AllTypes> HZ_COMPACT_SERIALIZER = new Serializer();

    private boolean mboolean;
    private boolean[] mbooleanArray;
    private byte mbyte;
    private byte[] mbyteArray;
    private short mshort;
    private short[] mshortArray;
    private int mint;
    private int[] mintArray;
    private long mlong;
    private long[] mlongArray;
    private float mfloat;
    private float[] mfloatArray;
    private double mdouble;
    private double[] mdoubleArray;
    private java.lang.String mstring;
    private java.lang.String[] mstringArray;
    private java.time.LocalDate mdate;
    private java.time.LocalDate[] mdateArray;
    private java.time.LocalTime mtime;
    private java.time.LocalTime[] mtimeArray;
    private java.time.LocalDateTime mtimestamp;
    private java.time.LocalDateTime[] mtimestampArray;
    private java.time.OffsetDateTime mtimestampWithTimezone;
    private java.time.OffsetDateTime[] mtimestampWithTimezoneArray;
    private Boolean mnullableboolean;
    private Boolean[] mnullablebooleanArray;
    private Byte mnullablebyte;
    private Byte[] mnullablebyteArray;
    private Short mnullableshort;
    private Short[] mnullableshortArray;
    private Integer mnullableint;
    private Integer[] mnullableintArray;
    private Long mnullablelong;
    private Long[] mnullablelongArray;
    private Float mnullablefloat;
    private Float[] mnullablefloatArray;
    private Double mnullabledouble;
    private Double[] mnullabledoubleArray;
    private NestedCompact mcompact;
    private NestedCompact[] mcompactArray;

    public AllTypes() {
    }

    public AllTypes(boolean mboolean, boolean[] mbooleanArray, byte mbyte, byte[] mbyteArray, short mshort, short[] mshortArray, int mint, int[] mintArray, long mlong, long[] mlongArray, float mfloat, float[] mfloatArray, double mdouble, double[] mdoubleArray, java.lang.String mstring, java.lang.String[] mstringArray, java.time.LocalDate mdate, java.time.LocalDate[] mdateArray, java.time.LocalTime mtime, java.time.LocalTime[] mtimeArray, java.time.LocalDateTime mtimestamp, java.time.LocalDateTime[] mtimestampArray, java.time.OffsetDateTime mtimestampWithTimezone, java.time.OffsetDateTime[] mtimestampWithTimezoneArray, Boolean mnullableboolean, Boolean[] mnullablebooleanArray, Byte mnullablebyte, Byte[] mnullablebyteArray, Short mnullableshort, Short[] mnullableshortArray, Integer mnullableint, Integer[] mnullableintArray, Long mnullablelong, Long[] mnullablelongArray, Float mnullablefloat, Float[] mnullablefloatArray, Double mnullabledouble, Double[] mnullabledoubleArray, NestedCompact mcompact, NestedCompact[] mcompactArray) {
        this.mboolean = mboolean;
        this.mbooleanArray = mbooleanArray;
        this.mbyte = mbyte;
        this.mbyteArray = mbyteArray;
        this.mshort = mshort;
        this.mshortArray = mshortArray;
        this.mint = mint;
        this.mintArray = mintArray;
        this.mlong = mlong;
        this.mlongArray = mlongArray;
        this.mfloat = mfloat;
        this.mfloatArray = mfloatArray;
        this.mdouble = mdouble;
        this.mdoubleArray = mdoubleArray;
        this.mstring = mstring;
        this.mstringArray = mstringArray;
        this.mdate = mdate;
        this.mdateArray = mdateArray;
        this.mtime = mtime;
        this.mtimeArray = mtimeArray;
        this.mtimestamp = mtimestamp;
        this.mtimestampArray = mtimestampArray;
        this.mtimestampWithTimezone = mtimestampWithTimezone;
        this.mtimestampWithTimezoneArray = mtimestampWithTimezoneArray;
        this.mnullableboolean = mnullableboolean;
        this.mnullablebooleanArray = mnullablebooleanArray;
        this.mnullablebyte = mnullablebyte;
        this.mnullablebyteArray = mnullablebyteArray;
        this.mnullableshort = mnullableshort;
        this.mnullableshortArray = mnullableshortArray;
        this.mnullableint = mnullableint;
        this.mnullableintArray = mnullableintArray;
        this.mnullablelong = mnullablelong;
        this.mnullablelongArray = mnullablelongArray;
        this.mnullablefloat = mnullablefloat;
        this.mnullablefloatArray = mnullablefloatArray;
        this.mnullabledouble = mnullabledouble;
        this.mnullabledoubleArray = mnullabledoubleArray;
        this.mcompact = mcompact;
        this.mcompactArray = mcompactArray;
    }

    public boolean getMboolean() {
        return mboolean;
    }

    public boolean[] getMbooleanArray() {
        return mbooleanArray;
    }

    public byte getMbyte() {
        return mbyte;
    }

    public byte[] getMbyteArray() {
        return mbyteArray;
    }

    public short getMshort() {
        return mshort;
    }

    public short[] getMshortArray() {
        return mshortArray;
    }

    public int getMint() {
        return mint;
    }

    public int[] getMintArray() {
        return mintArray;
    }

    public long getMlong() {
        return mlong;
    }

    public long[] getMlongArray() {
        return mlongArray;
    }

    public float getMfloat() {
        return mfloat;
    }

    public float[] getMfloatArray() {
        return mfloatArray;
    }

    public double getMdouble() {
        return mdouble;
    }

    public double[] getMdoubleArray() {
        return mdoubleArray;
    }

    public java.lang.String getMstring() {
        return mstring;
    }

    public java.lang.String[] getMstringArray() {
        return mstringArray;
    }

    public java.time.LocalDate getMdate() {
        return mdate;
    }

    public java.time.LocalDate[] getMdateArray() {
        return mdateArray;
    }

    public java.time.LocalTime getMtime() {
        return mtime;
    }

    public java.time.LocalTime[] getMtimeArray() {
        return mtimeArray;
    }

    public java.time.LocalDateTime getMtimestamp() {
        return mtimestamp;
    }

    public java.time.LocalDateTime[] getMtimestampArray() {
        return mtimestampArray;
    }

    public java.time.OffsetDateTime getMtimestampWithTimezone() {
        return mtimestampWithTimezone;
    }

    public java.time.OffsetDateTime[] getMtimestampWithTimezoneArray() {
        return mtimestampWithTimezoneArray;
    }

    public Boolean getMnullableboolean() {
        return mnullableboolean;
    }

    public Boolean[] getMnullablebooleanArray() {
        return mnullablebooleanArray;
    }

    public Byte getMnullablebyte() {
        return mnullablebyte;
    }

    public Byte[] getMnullablebyteArray() {
        return mnullablebyteArray;
    }

    public Short getMnullableshort() {
        return mnullableshort;
    }

    public Short[] getMnullableshortArray() {
        return mnullableshortArray;
    }

    public Integer getMnullableint() {
        return mnullableint;
    }

    public Integer[] getMnullableintArray() {
        return mnullableintArray;
    }

    public Long getMnullablelong() {
        return mnullablelong;
    }

    public Long[] getMnullablelongArray() {
        return mnullablelongArray;
    }

    public Float getMnullablefloat() {
        return mnullablefloat;
    }

    public Float[] getMnullablefloatArray() {
        return mnullablefloatArray;
    }

    public Double getMnullabledouble() {
        return mnullabledouble;
    }

    public Double[] getMnullabledoubleArray() {
        return mnullabledoubleArray;
    }

    public NestedCompact getMcompact() {
        return mcompact;
    }

    public NestedCompact[] getMcompactArray() {
        return mcompactArray;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;

        AllTypes that = (AllTypes) o;
        if (mboolean != that.mboolean) return false;
        if (!Arrays.equals(mbooleanArray, that.mbooleanArray)) return false;
        if (mbyte != that.mbyte) return false;
        if (!Arrays.equals(mbyteArray, that.mbyteArray)) return false;
        if (mshort != that.mshort) return false;
        if (!Arrays.equals(mshortArray, that.mshortArray)) return false;
        if (mint != that.mint) return false;
        if (!Arrays.equals(mintArray, that.mintArray)) return false;
        if (mlong != that.mlong) return false;
        if (!Arrays.equals(mlongArray, that.mlongArray)) return false;
        if (Float.compare(mfloat, that.mfloat) != 0) return false;
        if (!Arrays.equals(mfloatArray, that.mfloatArray)) return false;
        if (Double.compare(mdouble, that.mdouble) != 0) return false;
        if (!Arrays.equals(mdoubleArray, that.mdoubleArray)) return false;
        if (!Objects.equals(mstring, that.mstring)) return false;
        if (!Arrays.equals(mstringArray, that.mstringArray)) return false;
        if (!Objects.equals(mdate, that.mdate)) return false;
        if (!Arrays.equals(mdateArray, that.mdateArray)) return false;
        if (!Objects.equals(mtime, that.mtime)) return false;
        if (!Arrays.equals(mtimeArray, that.mtimeArray)) return false;
        if (!Objects.equals(mtimestamp, that.mtimestamp)) return false;
        if (!Arrays.equals(mtimestampArray, that.mtimestampArray)) return false;
        if (!Objects.equals(mtimestampWithTimezone, that.mtimestampWithTimezone)) return false;
        if (!Arrays.equals(mtimestampWithTimezoneArray, that.mtimestampWithTimezoneArray)) return false;
        if (!Objects.equals(mnullableboolean, that.mnullableboolean)) return false;
        if (!Arrays.equals(mnullablebooleanArray, that.mnullablebooleanArray)) return false;
        if (!Objects.equals(mnullablebyte, that.mnullablebyte)) return false;
        if (!Arrays.equals(mnullablebyteArray, that.mnullablebyteArray)) return false;
        if (!Objects.equals(mnullableshort, that.mnullableshort)) return false;
        if (!Arrays.equals(mnullableshortArray, that.mnullableshortArray)) return false;
        if (!Objects.equals(mnullableint, that.mnullableint)) return false;
        if (!Arrays.equals(mnullableintArray, that.mnullableintArray)) return false;
        if (!Objects.equals(mnullablelong, that.mnullablelong)) return false;
        if (!Arrays.equals(mnullablelongArray, that.mnullablelongArray)) return false;
        if (!Objects.equals(mnullablefloat, that.mnullablefloat)) return false;
        if (!Arrays.equals(mnullablefloatArray, that.mnullablefloatArray)) return false;
        if (!Objects.equals(mnullabledouble, that.mnullabledouble)) return false;
        if (!Arrays.equals(mnullabledoubleArray, that.mnullabledoubleArray)) return false;
        if (!Objects.equals(mcompact, that.mcompact)) return false;
        if (!Arrays.equals(mcompactArray, that.mcompactArray)) return false;

        return true;
    }

    @Override
    public int hashCode() {
        int result = 0;
        result = 31 * result + (mboolean ? 1 : 0);
        result = 31 * result + Arrays.hashCode(mbooleanArray);
        result = 31 * result + (int) mbyte;
        result = 31 * result + Arrays.hashCode(mbyteArray);
        result = 31 * result + (int) mshort;
        result = 31 * result + Arrays.hashCode(mshortArray);
        result = 31 * result + (int) mint;
        result = 31 * result + Arrays.hashCode(mintArray);
        result = 31 * result + (int) (mlong ^ (mlong >>> 32));
        result = 31 * result + Arrays.hashCode(mlongArray);
        result = 31 * result + (mfloat != +0.0f ? Float.floatToIntBits(mfloat) : 0);
        result = 31 * result + Arrays.hashCode(mfloatArray);
        long temp;
        temp = Double.doubleToLongBits(mdouble);
        result = 31 * result + (int) (temp ^ (temp >>> 32));
        result = 31 * result + Arrays.hashCode(mdoubleArray);
        result = 31 * result + Objects.hashCode(mstring);
        result = 31 * result + Arrays.hashCode(mstringArray);
        result = 31 * result + Objects.hashCode(mdate);
        result = 31 * result + Arrays.hashCode(mdateArray);
        result = 31 * result + Objects.hashCode(mtime);
        result = 31 * result + Arrays.hashCode(mtimeArray);
        result = 31 * result + Objects.hashCode(mtimestamp);
        result = 31 * result + Arrays.hashCode(mtimestampArray);
        result = 31 * result + Objects.hashCode(mtimestampWithTimezone);
        result = 31 * result + Arrays.hashCode(mtimestampWithTimezoneArray);
        result = 31 * result + Objects.hashCode(mnullableboolean);
        result = 31 * result + Arrays.hashCode(mnullablebooleanArray);
        result = 31 * result + Objects.hashCode(mnullablebyte);
        result = 31 * result + Arrays.hashCode(mnullablebyteArray);
        result = 31 * result + Objects.hashCode(mnullableshort);
        result = 31 * result + Arrays.hashCode(mnullableshortArray);
        result = 31 * result + Objects.hashCode(mnullableint);
        result = 31 * result + Arrays.hashCode(mnullableintArray);
        result = 31 * result + Objects.hashCode(mnullablelong);
        result = 31 * result + Arrays.hashCode(mnullablelongArray);
        result = 31 * result + Objects.hashCode(mnullablefloat);
        result = 31 * result + Arrays.hashCode(mnullablefloatArray);
        result = 31 * result + Objects.hashCode(mnullabledouble);
        result = 31 * result + Arrays.hashCode(mnullabledoubleArray);
        result = 31 * result + Objects.hashCode(mcompact);
        result = 31 * result + Arrays.hashCode(mcompactArray);

        return result;
    }

    @Override
    public String toString() {
        return "<AllTypes> {"
                + "mboolean=" + mboolean
                + ", mbooleanArray=" + Arrays.toString(mbooleanArray)
                + ", mbyte=" + mbyte
                + ", mbyteArray=" + Arrays.toString(mbyteArray)
                + ", mshort=" + mshort
                + ", mshortArray=" + Arrays.toString(mshortArray)
                + ", mint=" + mint
                + ", mintArray=" + Arrays.toString(mintArray)
                + ", mlong=" + mlong
                + ", mlongArray=" + Arrays.toString(mlongArray)
                + ", mfloat=" + mfloat
                + ", mfloatArray=" + Arrays.toString(mfloatArray)
                + ", mdouble=" + mdouble
                + ", mdoubleArray=" + Arrays.toString(mdoubleArray)
                + ", mstring=" + mstring
                + ", mstringArray=" + Arrays.toString(mstringArray)
                + ", mdate=" + mdate
                + ", mdateArray=" + Arrays.toString(mdateArray)
                + ", mtime=" + mtime
                + ", mtimeArray=" + Arrays.toString(mtimeArray)
                + ", mtimestamp=" + mtimestamp
                + ", mtimestampArray=" + Arrays.toString(mtimestampArray)
                + ", mtimestampWithTimezone=" + mtimestampWithTimezone
                + ", mtimestampWithTimezoneArray=" + Arrays.toString(mtimestampWithTimezoneArray)
                + ", mnullableboolean=" + mnullableboolean
                + ", mnullablebooleanArray=" + Arrays.toString(mnullablebooleanArray)
                + ", mnullablebyte=" + mnullablebyte
                + ", mnullablebyteArray=" + Arrays.toString(mnullablebyteArray)
                + ", mnullableshort=" + mnullableshort
                + ", mnullableshortArray=" + Arrays.toString(mnullableshortArray)
                + ", mnullableint=" + mnullableint
                + ", mnullableintArray=" + Arrays.toString(mnullableintArray)
                + ", mnullablelong=" + mnullablelong
                + ", mnullablelongArray=" + Arrays.toString(mnullablelongArray)
                + ", mnullablefloat=" + mnullablefloat
                + ", mnullablefloatArray=" + Arrays.toString(mnullablefloatArray)
                + ", mnullabledouble=" + mnullabledouble
                + ", mnullabledoubleArray=" + Arrays.toString(mnullabledoubleArray)
                + ", mcompact=" + mcompact
                + ", mcompactArray=" + Arrays.toString(mcompactArray)
                + '}';
    }

}