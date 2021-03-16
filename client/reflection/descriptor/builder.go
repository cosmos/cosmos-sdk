package descriptor

import (
	"errors"
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	ErrAlreadyExists = errors.New("descriptor: already exists")
)

func NewBuilder() *Builder {
	return &Builder{
		queriers:     newQueriers(),
		deliverables: newDeliverables(),
	}
}

// Builder builds a Chain descriptor
type Builder struct {
	queriers     queriers
	deliverables deliverables
}

func (b *Builder) Build() (Chain, error) {
	return newChain(b.queriers, b.deliverables), nil
}

func (b *Builder) RegisterDeliverable(desc protoreflect.MessageDescriptor) error {
	err := b.deliverables.insert(desc)
	if err != nil {
		return nil
	}
	return nil
}

func (b *Builder) RegisterQueryService(desc protoreflect.ServiceDescriptor) error {
	md := desc.Methods()
	for i := 0; i < md.Len(); i++ {
		method := md.Get(i)
		err := b.queriers.insert(desc, method)
		if err != nil {
			return err
		}
	}

	return nil
}

func newQueriers() queriers {
	return queriers{
		byName:   make(map[string]querier),
		byTMName: make(map[string]querier),
		byIndex:  nil,
	}
}

type queriers struct {
	byName   map[string]querier
	byTMName map[string]querier
	byIndex  []querier
}

func (q queriers) insert(sd protoreflect.ServiceDescriptor, md protoreflect.MethodDescriptor) error {
	name := (string)(md.FullName())
	if _, exists := q.byName[name]; exists {
		return fmt.Errorf("%s: %w", name, ErrAlreadyExists)
	}
	qr := newQuerier(sd, md)
	if _, exists := q.byTMName[qr.tmQueryPath]; exists {
		return fmt.Errorf("%s: %w", name, ErrAlreadyExists)
	}
	q.byName[name] = qr
	q.byTMName[qr.tmQueryPath] = qr
	q.byIndex = append(q.byIndex, qr)

	return nil
}

func (q queriers) Len() int {
	return len(q.byName)
}

func (q queriers) Get(i int) Querier {
	if i >= len(q.byIndex) {
		return nil
	}
	return q.byIndex[i]
}

func (q queriers) ByName(name string) Querier {
	if o, exists := q.byName[name]; exists {
		return o
	}
	return nil
}

func (q queriers) ByTMName(tmName string) Querier {
	if o, exists := q.byTMName[tmName]; exists {
		return o
	}
	return nil
}

func newQuerier(sd protoreflect.ServiceDescriptor, md protoreflect.MethodDescriptor) querier {
	return querier{
		desc:        md,
		tmQueryPath: fmt.Sprintf("/%s/%s", sd.FullName(), md.Name()), // TODO: why in the sdk we broke standard grpc query method invocation naming?
	}
}

type querier struct {
	desc        protoreflect.MethodDescriptor
	tmQueryPath string
}

func (q querier) Descriptor() protoreflect.MethodDescriptor {
	return q.desc
}

func (q querier) TMQueryPath() string {
	return q.tmQueryPath
}

func newDeliverables() deliverables {
	return deliverables{
		byName:  make(map[string]deliverable),
		byIndex: nil,
	}
}

type deliverables struct {
	byName  map[string]deliverable
	byIndex []deliverable
}

func (d deliverables) insert(md protoreflect.MessageDescriptor) error {
	name := (string)(md.FullName())
	if _, exists := d.byName[name]; exists {
		return fmt.Errorf("%w: %s", ErrAlreadyExists, name)
	}

	deliverable := newDeliverable(md)
	d.byName[name] = deliverable
	d.byIndex = append(d.byIndex, deliverable)

	return nil
}

func (d deliverables) Len() int {
	return len(d.byName)
}

func (d deliverables) Get(i int) Deliverable {
	if i >= len(d.byIndex) {
		return nil
	}
	return d.byIndex[i]
}

func (d deliverables) ByName(name string) Deliverable {
	desc, exists := d.byName[name]
	if !exists {
		return nil
	}
	return desc
}

func newDeliverable(desc protoreflect.MessageDescriptor) deliverable {
	return deliverable{desc: desc}
}

type deliverable struct {
	desc protoreflect.MessageDescriptor
}

func (d deliverable) Descriptor() protoreflect.MessageDescriptor {
	return d.desc
}

func newChain(q queriers, d deliverables) chain {
	return chain{
		q: q,
		d: d,
	}
}

type chain struct {
	q queriers
	d deliverables
}

func (c chain) Config() Config {
	panic("implement me")
}

func (c chain) Deliverables() Deliverables {
	return c.d
}

func (c chain) Queriers() Queriers {
	return c.q
}
