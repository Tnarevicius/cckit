package serialize_test

import (
	"testing"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/hyperledger-labs/cckit/serialize"
	"github.com/hyperledger-labs/cckit/serialize/testdata"
	"github.com/hyperledger-labs/cckit/testing/gomega"
)

func TestState(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "State suite")
}

var (
	StringToSerialize = `my-string`
	BytesToSerialize  = []byte(`some bytes`)
	ProtoToSerialize  = &testdata.Payment{
		Type:         "some-type",
		Id:           "some-id",
		Amount:       100,
		Key:          []byte("public-key"),
		Deadline:     timestamppb.New(time.Now().AddDate(0, 0, 2)),
		SnakeOrCamel: "snake_case or camelCase",
	}
	err error
)

var _ = Describe(`Generic serializer`, func() {

	serializer := serialize.DefaultSerializer

	Context(`Bool`, func() {
		var bTrue, bFalse []byte
		It(`serialize`, func() {
			bTrue, err = serializer.ToBytesFrom(true)
			Expect(err).NotTo(HaveOccurred())
			Expect(bTrue).To(Equal([]byte(`true`)))

			bFalse, err = serializer.ToBytesFrom(false)
			Expect(err).NotTo(HaveOccurred())
			Expect(bFalse).To(Equal([]byte(`false`)))
		})

		It(`deserialize`, func() {
			eTrue, err := serializer.FromBytesTo(bTrue, serialize.TypeBool)
			Expect(err).NotTo(HaveOccurred())
			Expect(eTrue.(bool)).To(Equal(true))

			eFalse, err := serializer.FromBytesTo(bFalse, serialize.TypeBool)
			Expect(err).NotTo(HaveOccurred())
			Expect(eFalse.(bool)).To(Equal(false))
		})

	})

	Context(`String`, func() {
		It(`Serialize`, func() {
			bStr, err := serializer.ToBytesFrom(StringToSerialize)
			Expect(err).NotTo(HaveOccurred())
			Expect(bStr).To(Equal([]byte(StringToSerialize)))

			eStr, err := serializer.FromBytesTo(bStr, serialize.TypeString)
			Expect(err).NotTo(HaveOccurred())
			Expect(eStr.(string)).To(Equal(StringToSerialize))
		})
	})

	Context(`Nil`, func() {
		It(`Serialize`, func() {
			bNil, err := serializer.ToBytesFrom(nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(bNil).To(Equal([]byte{}))
		})
	})

	Context(`Bytes`, func() {

		var serializedBytes1 []byte
		It(`serialize`, func() {
			serializedBytes1, err = serializer.ToBytesFrom(BytesToSerialize)
			Expect(err).NotTo(HaveOccurred())
			Expect(serializedBytes1).To(Equal(BytesToSerialize))
		})

		It(`deserialize`, func() {
			deserializedBytes, err := serializer.FromBytesTo(serializedBytes1, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(deserializedBytes).To(Equal(BytesToSerialize))

			deserializedBytes2, err := serializer.FromBytesTo(serializedBytes1, []byte{})
			Expect(err).NotTo(HaveOccurred())
			Expect(deserializedBytes2).To(Equal(BytesToSerialize))
		})
	})

	Context(`Proto`, func() {

		var serializedProto1 []byte
		It(`serialize`, func() {
			serializedProto1, err = serializer.ToBytesFrom(ProtoToSerialize)
			Expect(err).NotTo(HaveOccurred())

			bb, err := proto.Marshal(ProtoToSerialize)
			Expect(err).NotTo(HaveOccurred())

			Expect(serializedProto1).To(Equal(bb))
		})

		It(`deserialize`, func() {
			deserializedProto, err := serializer.FromBytesTo(serializedProto1, &testdata.Payment{})
			Expect(err).NotTo(HaveOccurred())
			Expect(deserializedProto).To(gomega.StringerEqual(ProtoToSerialize))
		})
	})
})

var _ = Describe(`Prefer JSON serializer`, func() {

	serializer := serialize.PreferJSONSerializer

	var serializedProtoAsJSON []byte
	It(`serialize (snake_case)`, func() {
		serializedProtoAsJSON, err = serializer.ToBytesFrom(ProtoToSerialize)
		Expect(err).NotTo(HaveOccurred())

		bb, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(ProtoToSerialize)
		Expect(err).NotTo(HaveOccurred())

		Expect(serializedProtoAsJSON).To(Equal(bb))
	})

	It(`deserialize (snake_case)`, func() {
		deserializedProto, err := serializer.FromBytesTo(serializedProtoAsJSON, &testdata.Payment{})
		Expect(err).NotTo(HaveOccurred())
		Expect(deserializedProto).To(gomega.StringerEqual(ProtoToSerialize))
	})

	It(`serialize (camelCase)`, func() {
		serializer.UseProtoNames = false

		serializedProtoAsJSON, err = serializer.ToBytesFrom(ProtoToSerialize)
		Expect(err).NotTo(HaveOccurred())

		bb, err := protojson.Marshal(ProtoToSerialize)
		Expect(err).NotTo(HaveOccurred())

		Expect(serializedProtoAsJSON).To(Equal(bb))
	})

	It(`deserialize (camelCase)`, func() {
		deserializedProto, err := serializer.FromBytesTo(serializedProtoAsJSON, &testdata.Payment{})
		Expect(err).NotTo(HaveOccurred())
		Expect(deserializedProto).To(gomega.StringerEqual(ProtoToSerialize))
	})
})
