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

func NameTemplate(name string, iconId string, pronoun PronounType, isProperName bool) *entity.DefaultTemplate {
	return entity.NewDefaultTemplate((*Name)(nil), NameComponent, map[string]interface{}{
		"Name":         name,
		"IconId":       iconId,
		"Pronoun":      pronoun,
		"IsProperName": isProperName,
	})
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
//   thename's: The possessive form of the entity's name with definite article
//   aname's: The possessive form of the entity's name with indefinite article
//   s: Trailing 's' for verbs ("The goblin hitS") for others, empty for second person.
//   pronoun: The pronoun used to refer to the entity. 'You' for second person.
//   pronoun's: Possessive form of the pronoun.
//   self: Reflective pronoun for the entity.
//   is: 'is' for others, 'are' for second person.
//   accusative: Accusative case pronoun you, him, her, it
//
// In addition the value
//
//   you
//
// evaluates to true for second person and false for others. Use it like this:
//
//   "{sub.Thename} {.section sub.you}miss (SECOND PERSON){.or}misses (OTHER){.end} {obj.thename}..."
//
// Although for the common are/is case, there's the ".is" shorthand.
func EMsg(format string, subjectId, objectId entity.Id, a ...interface{}) {
	str := fmt.Sprintf(format, a)
	str = FormatMessage(str, subjectId, objectId)
	Fx().Print(str)
}

func FormatMessage(fmtStr string, subjectId, objectId entity.Id) string {
	subjectWords := entityTemplateWorlds(subjectId)
	objectWords := entityTemplateWorlds(objectId)

	mp := map[string]interface{}{"sub": subjectWords, "obj": objectWords}

	// XXX: A chance to optimize: Cache the parsed fmtStr templates somewhere
	// at this point.
	tmpl := template.MustParse(fmtStr, nil)

	buffer := new(bytes.Buffer)
	tmpl.Execute(mp, buffer)
	return buffer.String()
}

func entityTemplateWorlds(id entity.Id) map[string]string {
	if id == PlayerId() {
		return YouTemplateWords()
	}

	name := GetNameComp(id)
	if name == nil {
		return make(map[string]string)
	}

	return name.TemplateWords()
}


// GetCapName returns the capitalized name of an entity.
func GetCapName(id entity.Id) string { return txt.Capitalize(GetName(id)) }

func (self *Name) TemplateWords() (result map[string]string) {
	result = make(map[string]string)
	result["name"] = self.Name
	result["you"] = "" // Empty string acts as 'false', use in conditionals.

	aArticle := txt.GuessIndefiniteArticle(self.Name)

	if self.IsProperName {
		result["thename"] = self.Name
		result["aname"] = self.Name
		result["thename's"] = self.Name + "'s"
		result["aname's"] = self.Name + "'s"
	} else {
		result["thename"] = "the " + self.Name
		result["aname"] = aArticle + " " + self.Name
		result["thename's"] = "the " + self.Name + "'s"
		result["aname's"] = aArticle + " " + self.Name + "'s"
	}

	result["is"] = "is"

	result["name's"] = self.Name + "'s"

	result["s"] = "s"
	switch self.Pronoun {
	case PronounHe:
		result["pronoun"] = "he"
		result["pronoun's"] = "his"
		result["self"] = "himself"
		result["accusative"] = "him"
	case PronounShe:
		result["pronoun"] = "she"
		result["pronoun's"] = "her"
		result["self"] = "herself"
		result["accusative"] = "her"
	case PronounThey:
		result["pronoun"] = "they"
		result["pronoun's"] = "their"
		result["self"] = "themself"
		result["accusative"] = "them"
	default:
		result["pronoun"] = "it"
		result["pronoun's"] = "its"
		result["self"] = "itself"
		result["accusative"] = "it"
	}

	addCapitalizedFields(result)

	return result
}

func YouTemplateWords() (result map[string]string) {
	result = map[string]string{
		"you": "1", // Acts as 'true', use in conditionals.

		"name":       "you",
		"thename":    "you",
		"aname":      "you",
		"name's":     "your",
		"thename's":  "your",
		"aname's":    "your",
		"s":          "",
		"pronoun":    "you",
		"pronoun's":  "your",
		"self":       "yourself",
		"is":         "are",
		"accusative": "you",
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
