package test;

import com.hazelcast.nio.serialization.compact.CompactReader;
import com.hazelcast.nio.serialization.compact.CompactSerializer;
import com.hazelcast.nio.serialization.compact.CompactWriter;

import javax.annotation.Nonnull;
import java.util.Arrays;
import java.util.Objects;

public class TypesWithDefaults {

    public static final class Serializer implements CompactSerializer<TypesWithDefaults> {
        @Nonnull
        @Override
        public TypesWithDefaults read(@Nonnull CompactReader reader) {
            boolean mboolean = reader.readBoolean("mboolean", true);
            byte mbyte = reader.readInt8("mbyte", (byte) 8);
            short mshort = reader.readInt16("mshort", (short) 16);
            int mint = reader.readInt32("mint", 32);
            long mlong = reader.readInt64("mlong", 64);
            float mfloat = reader.readFloat32("mfloat", (float) 32.32);
            double mdouble = reader.readFloat64("mdouble", 64.64);
            return new TypesWithDefaults(mboolean, mbyte, mshort, mint, mlong, mfloat, mdouble);
        }

        @Override
        public void write(@Nonnull CompactWriter writer, @Nonnull TypesWithDefaults object) {
            writer.writeBoolean("mboolean", object.mboolean);
            writer.writeInt8("mbyte", object.mbyte);
            writer.writeInt16("mshort", object.mshort);
            writer.writeInt32("mint", object.mint);
            writer.writeInt64("mlong", object.mlong);
            writer.writeFloat32("mfloat", object.mfloat);
            writer.writeFloat64("mdouble", object.mdouble);    
        }
    };
    
    
    public static final CompactSerializer<TypesWithDefaults> HZ_COMPACT_SERIALIZER = new Serializer();

    private boolean mboolean = true;
    private byte mbyte = 8;
    private short mshort = 16;
    private int mint = 32;
    private long mlong = 64;
    private float mfloat = (float) 32.32;
    private double mdouble = 64.64;

    public TypesWithDefaults() {
    }

    public TypesWithDefaults(boolean mboolean, byte mbyte, short mshort, int mint, long mlong, float mfloat, double mdouble) {
        this.mboolean = mboolean;
        this.mbyte = mbyte;
        this.mshort = mshort;
        this.mint = mint;
        this.mlong = mlong;
        this.mfloat = mfloat;
        this.mdouble = mdouble;
    }
    
    public boolean getMboolean() {
        return mboolean;
    }

    public byte getMbyte() {
        return mbyte;
    }

    public short getMshort() {
        return mshort;
    }

    public int getMint() {
        return mint;
    }

    public long getMlong() {
        return mlong;
    }

    public float getMfloat() {
        return mfloat;
    }

    public double getMdouble() {
        return mdouble;
    }   
    
    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;

        TypesWithDefaults that = (TypesWithDefaults) o;
        if (mboolean != that.mboolean) return false;
        if (mbyte != that.mbyte) return false;
        if (mshort != that.mshort) return false;
        if (mint != that.mint) return false;
        if (mlong != that.mlong) return false;
        if (Float.compare(mfloat, that.mfloat) != 0) return false;
        if (Double.compare(mdouble, that.mdouble) != 0) return false;
        return true;
    }
    
    @Override
    public int hashCode() {
        int result = 0;
        result = 31 * result + (mboolean ? 1 : 0);
        result = 31 * result + (int) mbyte;
        result = 31 * result + (int) mshort;
        result = 31 * result + (int) mint;
        result = 31 * result + (int) (mlong ^ (mlong >>> 32));
        result = 31 * result + (mfloat != +0.0f ? Float.floatToIntBits(mfloat) : 0);
        long temp;
        temp = Double.doubleToLongBits(mdouble);
        result = 31 * result + (int) (temp ^ (temp >>> 32));  
        return result;      
    }
    
    @Override
    public String toString() {
        return "<TypesWithDefaults> {"
                + ", + mboolean=" + mboolean
                + ", + mbyte=" + mbyte
                + ", + mshort=" + mshort
                + ", + mint=" + mint
                + ", + mlong=" + mlong
                + ", + mfloat=" + mfloat
                + ", + mdouble=" + mdouble          
                + '}';
    }       
     
}