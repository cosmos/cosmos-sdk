package desc

import (
	"google.golang.org/protobuf/reflect/protoreflect"
)

// HasKinds recursively checks if the given Message Descriptor contains the provided field kinds.
func HasKinds(desc protoreflect.MessageDescriptor, kinds ...protoreflect.Kind) bool {
	for _, fKind := range kinds {
		if HasKind(desc, fKind) {
			return true
		}
	}
	return false
}

// HasKind recursively checks if the given MessageDescriptor contains the provided field kind.
func HasKind(desc protoreflect.MessageDescriptor, fKind protoreflect.Kind) bool {
	for i := 0; i < desc.Fields().Len(); i++ {
		fd := desc.Fields().Get(i)
		switch {
		case fd.IsMap():
			keyKind := fd.MapKey().Kind()
			valueKind := fd.MapValue().Kind()
			if keyKind == fKind || valueKind == fKind {
				return true
			}
			if valueKind == protoreflect.MessageKind {
				if HasKind(fd.MapValue().Message(), fKind) {
					return true
				}
			}
		case fd.Kind() == fKind:
			return true
		}
	}

	return false
}

// HasMap returns true if the given message descriptor
// contains a map field. It inspects the message recursively.
func HasMap(desc protoreflect.MessageDescriptor) bool {
	return hasMap(map[protoreflect.FullName]struct{}{}, desc)
}

func hasMap(inspected map[protoreflect.FullName]struct{}, desc protoreflect.MessageDescriptor) bool {

	var toInspect []protoreflect.MessageDescriptor

	for i := 0; i < desc.Fields().Len(); i++ {
		fd := desc.Fields().Get(i)
		switch {
		case fd.IsMap():
			return true
		case fd.Kind() == protoreflect.MessageKind:
			toInspect = append(toInspect, fd.Message())
		default:
			continue
		}
	}

	inspected[desc.FullName()] = struct{}{}

	for _, insp := range toInspect {
		if _, ok := inspected[desc.FullName()]; ok {
			continue
		}
		if hasMap(inspected, insp) {
			return true
		}
		inspected[insp.FullName()] = struct{}{}
	}

	return false
}

// Dependencies returns the dependency list of the given protoreflect.FileDescriptor
// recursively and in order of required registration.
// It assumes the protoreflect.FileDescriptor is correctly formed.
func Dependencies(desc protoreflect.FileDescriptor) []protoreflect.FileDescriptor {
	return dependencies(map[string]struct{}{}, desc)
}

func dependencies(knownDeps map[string]struct{}, desc protoreflect.FileDescriptor) []protoreflect.FileDescriptor {
	var deps []protoreflect.FileDescriptor

	for i := 0; i < desc.Imports().Len(); i++ {
		dep := desc.Imports().Get(i)
		// if it's known we skip the parsing
		if _, known := knownDeps[dep.Path()]; known {
			continue
		}

		// get the dependency's dependencies
		depDeps := dependencies(knownDeps, dep)
		deps = append(deps, depDeps...)
		deps = append(deps, dep)

		knownDeps[dep.Path()] = struct{}{}
	}

	return deps
}
