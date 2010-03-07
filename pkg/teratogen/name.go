package teratogen

import (
	"bytes"
	"fmt"
	"hyades/entity"
	"hyades/txt"
	"template"
	"unicode"
)

const NameComponent = entity.ComponentFamily("name")


type NameTemplate Name

func (self NameTemplate) Derive(c entity.ComponentTemplate) entity.ComponentTemplate {
	return c
}

func (self NameTemplate) MakeComponent(manager *entity.Manager, guid entity.Id) {
	manager.Handler(NameComponent).Add(guid, &Name{self.Name, self.IconId, self.Pronoun, self.IsProperName})
}


type PronounType int

const (
	PronounIt = PronounType(iota)
	PronounHe
	PronounShe
	// Use 'they' for hive minds and as gender-neutral personal pronoun.
	PronounThey
)

// Name component.
type Name struct {
	Name         string
	IconId       string
	Pronoun      PronounType
	IsProperName bool
}


// GetName returns the name of an entity. If the entity has no name component,
// it returns a string representation of its id value.
func GetName(id entity.Id) string {
	if nameComp := GetManager().Handler(NameComponent).Get(id); nameComp != nil {
		return nameComp.(*Name).Name
	}
	return string(id)
}

func GetNameComp(id entity.Id) *Name {
	if nameComp := GetManager().Handler(NameComponent).Get(id); nameComp != nil {
		return nameComp.(*Name)
	}
	return nil
}

func GetIconId(id entity.Id) string {
	if nameComp := GetManager().Handler(NameComponent).Get(id); nameComp != nil {
		return nameComp.(*Name).IconId
	}
	return ""
}


func Msg(format string, a ...interface{}) { Fx().Print(fmt.Sprintf(format, a)) }

// EMsg is a message of an entity doing something to another entity. It uses
// template formatting after the Sprintf phase. The format strings are of the type
//
//   {sub.name}
//
// for referring to subject, and
//
//   {obj.name}
//
// for referring to the object. You can capitalize the field name to get a
// capitalized value:
//
//   {obj.name} => "goblin"
//   {obj.Name} => "Goblin"
//
// Either the subject or the object may be the player, in which case the
// second person tense ('you') is used.
//
// The list of format values for entities:
//
//   name: The name of the entity without an article
//   thename: The name of the entity with a definite article, unless it's a proper name
//   aname: The name of the entity with an indefinite article, unless it's a proper name
//   name's: The possessive form of the entity's name
//   s: Trailing 's' for verbs ("The goblin hitS") for others, empty for second person.
//   pronoun: The pronoun used to refer to the entity. 'You' for second person.
//   pronoun's: Possessive form of the pronoun.
//   self: Reflective pronoun for the entity.

func EMsg(format string, subjectId, objectId entity.Id, a ...interface{}) {
	str := fmt.Sprintf(format, a)
	str = FormatMessage(str, subjectId, objectId)
	Fx().Print(str)
}

func FormatMessage(fmtStr string, subjectId, objectId entity.Id) string {
	subjectName, objectName := GetNameComp(subjectId), GetNameComp(objectId)

	var subjectWords, objectWords map[string]string

	if subjectId == PlayerId() {
		subjectWords = YouTemplateWords()
	} else {
		subjectWords = subjectName.TemplateWords()
	}

	if objectId == PlayerId() {
		objectWords = YouTemplateWords()
	} else {
		objectWords = objectName.TemplateWords()
	}

	mp := map[string]interface{}{"sub": subjectWords, "obj": objectWords}

	// XXX: A chance to optimize: Cache the parsed fmtStr templates somewhere
	// at this point.
	tmpl := template.MustParse(fmtStr, nil)

	buffer := new(bytes.Buffer)
	tmpl.Execute(mp, buffer)
	return buffer.String()
}

// GetCapName returns the capitalized name of an entity.
func GetCapName(id entity.Id) string { return txt.Capitalize(GetName(id)) }

func (self *Name) TemplateWords() (result map[string]string) {
	result = make(map[string]string)
	result["name"] = self.Name
	if self.IsProperName {
		result["thename"] = self.Name
		result["aname"] = self.Name
	} else {
		result["thename"] = "the " + self.Name
		result["aname"] = "a " + self.Name
	}

	result["name's"] = self.Name + "'s"
	result["s"] = "s"
	switch self.Pronoun {
	case PronounHe:
		result["pronoun"] = "he"
		result["pronoun's"] = "his"
		result["self"] = "himself"
	case PronounShe:
		result["pronoun"] = "she"
		result["pronoun's"] = "her"
		result["self"] = "herself"
	case PronounThey:
		result["pronoun"] = "they"
		result["pronoun's"] = "their"
		result["self"] = "themself"
	default:
		result["pronoun"] = "it"
		result["pronoun's"] = "its"
		result["self"] = "itself"
	}

	addCapitalizedFields(result)

	return result
}

func YouTemplateWords() (result map[string]string) {
	result = map[string]string{
		"you": "1", // Set to defined, use this as a conditional.

		"name":      "you",
		"thename":   "you",
		"aname":     "you",
		"name's":    "your",
		"s":         "",
		"pronoun":   "you",
		"pronoun's": "your",
		"self":      "yourself",
	}

	addCapitalizedFields(result)

	return result
}

// addCapitalizedFields adds a words["Foo"] = "Bar" for every words["foo"] = "bar".
func addCapitalizedFields(words map[string]string) {
	for k, v := range words {
		// XXX: Is straight string indexing the right way to get at unicode
		// chars?
		if unicode.IsLower(int(k[0])) {
			words[txt.Capitalize(k)] = txt.Capitalize(v)
		}
	}
}
