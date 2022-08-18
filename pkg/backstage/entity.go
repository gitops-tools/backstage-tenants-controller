package backstage

import "encoding/json"

// entity represents a Backstage entity from the entities API endpoint.
//
// https://backstage.io/docs/features/software-catalog/descriptor-format#overall-shape-of-an-entity
type entity struct {
	TypeMeta  `json:",inline"`
	Metadata  Metadata   `json:"metadata,omitempty"`
	Relations []Relation `json:"relations,omitempty"`

	Spec json.RawMessage
}

// Metadata is the metadata for Backstage entities.
//
// https://backstage.io/docs/features/software-catalog/descriptor-format#common-to-all-kinds-the-metadata
type Metadata struct {
	// Name of the entity.
	Name string `json:"name"`

	// Namespace is optional ID of a namespace that the entity belongs to.
	Namespace string `json:"namespace,omitempty"`

	// Description is a human readable description of the entity, to be shown in Backstage.
	Description string `json:"description,omitempty"`

	// Title is the display name of the entity, to be presented in user interfaces instead of the name property above, when available.
	Title string `json:"title,omitempty"`

	// 	Labels are optional key/value pairs of that are attached to the entity, and their use is identical to Kubernetes object labels.
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations are arbitrary non-identifying metadata attached to the entity, identical in use to Kubernetes object annotations.
	Annotations map[string]string `json:"annotations,omitempty"`

	// Tags are a list of single-valued strings, for example to classify catalog entities in various ways.
	Tags []string `json:"tags,omitempty"`

	// Links are a list of external hyperlinks related to the entity.
	Links []Link `json:"links,omitempty"`

	// UID is the unique ID assigned to this resource.
	UID string `json:"uid"`
}

// Link can provide additional contextual information that may be located
// outside of Backstage itself.
//
// https://backstage.io/docs/features/software-catalog/descriptor-format#links-optional
type Link struct {
	// URL is standard URI format.
	URL string `json:"url"`

	// Title is a user friendly display name for the link.
	Title string `json:"title,omitempty"`

	// Icon is a key representing a visual icon to be displayed in the UI.
	Icon string `json:"icon,omitempty"`

	// Type is an optional value to categorize links into specific groups.
	Type string `json:"type,omitempty"`
}

// TypeMeta is the entity envelope.
//
// https://backstage.io/docs/features/software-catalog/descriptor-format#common-to-all-kinds-the-envelope
type TypeMeta struct {
	// Kind is the high level entity type being described.
	Kind string `json:"kind,omitempty"`
	// APIVersion is the version of specification format for that particular entity that the specification is made against.
	APIVersion string `json:"apiVersion,omitempty"`
}

// Relation is a relation from an entity to another entity.
//
// https://backstage.io/docs/features/software-catalog/descriptor-format#common-to-all-kinds-relations
type Relation struct {
	// The type of relation FROM a source entity TO the target entity.
	Type string `json:"type"`

	// TargetRef is a string version of the Target.
	TargetRef string `json:"targetRef,omitempty"`

	// Target is a complete compound reference to the other end of the relation.
	Target struct {
		// Name is the name of the entity being referenced.
		Name string `json:"name"`
		// Namespace is the namespace of the entity being referenced.
		Namespace string `json:"namespace"`
		// Kind is the kind of entity being referenced.
		Kind string `json:"kind"`
	} `json:"target"`
}
