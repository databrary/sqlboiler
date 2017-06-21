# Databrary Changes

This documentation outlines changes made to [SQLBoiler](https://github.com/vattle/sqlboiler) in order to support
custom types (i.e. project specific database types) and views.

## Table of Contents



* [Type mapping](#type-mapping)
* [Templates](#templates)
 * [Custom Types](#custom-types)
    * [Testing](#testing)
    * [Imports](#imports)
 * [Views](#views)
* [Tests](#tests)


## Type mapping

[SQLBoiler](https://github.com/vattle/sqlboiler) works by inspecting all of the columns in all of the tables in a schema
and mapping their types (int, float, varchar, enum, etc.) to go types. For most of
the basic types SQLBoiler maps correctly (for nullable types it uses https://github.com/nullbio/null)

The mapping happens in the database drivers - for databrary it happens in the postgres driver @ `sqlboiler/bdb/drivers/postgres.go`. The
key piece of code is the sql query that returns the metadata for the columns of a table in a schema:

```sql
select
    c.column_name,
    (
        case when pgt.typtype = 'e'
        then
        (
            select 'enum.' || c.udt_name || '(''' || string_agg(labels.label, ''',''') || ''')'
            from (
                select pg_enum.enumlabel as label
                from pg_enum
                where pg_enum.enumtypid =
                (
                    select typelem
                    from pg_type
                    where pg_type.typtype = 'b' and pg_type.typname = ('_' || c.udt_name)
                    limit 1
                )
                order by pg_enum.enumsortorder
            ) as labels
        )
        else c.data_type
        end
    ) as column_type,

    c.udt_name,
    e.data_type as array_type,
    c.column_default,

    c.is_nullable = 'YES' as is_nullable,
    (select exists(
        select 1
        from information_schema.table_constraints tc
        inner join information_schema.constraint_column_usage as ccu on tc.constraint_name = ccu.constraint_name
        where tc.table_schema = 'public' and tc.constraint_type = 'UNIQUE' and ccu.constraint_schema = 'public' and ccu.table_name = c.table_name and ccu.column_name = c.column_name and
            (select count(*) from information_schema.constraint_column_usage where constraint_schema = 'public' and constraint_name = tc.constraint_name) = 1
    )) OR
    (select exists(
        select 1
        from pg_indexes pgix
        inner join pg_class pgc on pgix.indexname = pgc.relname and pgc.relkind = 'i' and pgc.relnatts = 1
        inner join pg_index pgi on pgi.indexrelid = pgc.oid
        inner join pg_attribute pga on pga.attrelid = pgi.indrelid and pga.attnum = ANY(pgi.indkey)
        where
            pgix.schemaname = 'public' and pgix.tablename = c.table_name and pga.attname = c.column_name and pgi.indisunique = true
    )) as is_unique

    from information_schema.columns as c
    inner join pg_namespace as pgn on pgn.nspname = c.udt_schema
    left join pg_type pgt on c.data_type = 'USER-DEFINED' and pgn.oid = pgt.typnamespace and c.udt_name = pgt.typname
    left join information_schema.element_types e
        on ((c.table_catalog, c.table_schema, c.table_name, 'TABLE', c.dtd_identifier)
        = (e.object_catalog, e.object_schema, e.object_name, e.object_type, e.collection_type_identifier))
    where c.table_name = 'slot_release' and c.table_schema = 'public';
```

You don't need to understand how this works really (but I have fixed a bug in it before).
The above query returns the metadata for the columns of `slot_release`. It's a good example because it includes
a basic type, an enum, and a custom type. The output is

<br />

column_name|column_type|udt_name|array_type|column_default|is_nullable|is_unique
------------- | ------------- | ------------- | ------------- | ------------- | ------------- | -------------
container | integer | int4 | | | f | f
segment|USER-DEFINED|segment| | | f | f
release|enum.release('PRIVATE','SHARED','EXCERPTS','PUBLIC')|release| | | f | f

SQLBoiler will correct produce a struct representation of `slot_release`:

```go
type SlotRelease struct {
    Container int                  `db:"container" json:"slotRelease_container"`
    Segment   custom_types.Segment `db:"segment" json:"slotRelease_segment"`
    Release   custom_types.Release `db:"release" json:"slotRelease_release"`

    R *slotReleaseR `db:"-" json:"-"`
    L slotReleaseL  `db:"-" json:"-"`
}
```

Don't pay attention to the `` markup yet (https://stackoverflow.com/a/30889373). Note that `container` in `slot_release` is mapped to
`Container int` in `SlotRelease` - the standard distribution of SQLBoiler would do this successfully
too since `container` is of `int` type. Now note that `segment` in `slot_release` has a `column_type` of `USER-DEFINED`.
In the standard distribution this would lead to `Segment` being of `string` type in `SlotRelease`. It's this extension
that I've implemented:

```go
case "USER-DEFINED":
    if c.UDTName == "hstore" {
        c.Type = "types.HStore"
        c.DBType = "hstore"
    } else {
        c.Type = "custom_types." + strings.Join(strings.Split(strings.Title(strings.Join(strings.Split(c.UDTName, "_"), " ")), " "), "")
        fmt.Printf("Warning: USER-DEFINED data type detected: %s %s\n", c.UDTName, c.Name)
        fmt.Printf("used %s \n", c.Type)
    }
    c.IsCustom = true
```

in particular this line

```go
c.Type = "custom_types." + strings.Join(strings.Split(strings.Title(strings.Join(strings.Split(c.UDTName, "_"), " ")), " "), "")
```

assigns type `custom_types.AliceBobFrank` to a column with a `udt_name` of `alice_bob_frank`. `c.IsCustom = true` identifies
the column as a custom column for further along in the transformation pipeline.

The `release` enum column is supported by this case statement

```go
case "release":
    c.Type = "custom_types.Release"
    c.IsCustom = true
```

I've also implemented support for `inet` and `interval` types.

**All of these type identifiers have corresponding implementations in `databrary/db/models/custom_types`**
hence the prefix `custom_types` is significant: it identifies the module that will need to be imported
in the model source for models that have "custom types".

Almost all of the other code differences are purely to support these custom types in the templates.

## Templates

The stock templates distributed with SQLBoiler have been heavily (heavily!) modified. They're
all in `databrary/config/sqlboiler/models/sqlboiler_models`

### Custom Types

#### Testing

Most of the work done to extend the templates to incorporate custom types went into
incorporating them into the tests.

This function

```go
bdb/Column.go

func FilterColumnsByCustom(columns []Column) []Column {
    var cols []Column

    for _, c := range columns {
        if c.IsCustom {
            cols = append(cols, c)
        }
    }
    return cols
}
```
is mapped to `filterColumnsByCustom` @ `sqlboiler/boilingcore/templates.go` and used here

`{{$varNameSingular}}ColumnsWithCustom = []string{{"{"}}{{.Table.Columns | filterColumnsByCustom | columnNames | stringMap .StringFuncs.quoteWrap | join ","}}{{"}"}}`

to identify the fields/columns that correspond custom db types.

For example in `testAccountToManyWhoTagUses`

```go
foreignBlacklist := tagUseColumnsWithDefault
foreignBlacklist = append(foreignBlacklist, tagUseColumnsWithCustom...)

if err := randomize.Struct(seed, &b, tagUseDBTypes, false, foreignBlacklist...); err != nil {
    t.Errorf("Unable to randomize TagUse struct: %s", err)
}
if err := randomize.Struct(seed, &c, tagUseDBTypes, false, foreignBlacklist...); err != nil {
    t.Errorf("Unable to randomize TagUse struct: %s", err)
}
b.Segment = custom_types.SegmentRandom()
c.Segment = custom_types.SegmentRandom()
```

In particular note that `tagUseColumnsWithDefault` is included in the list
columns that are "blacklisted" from `randomize.Struct`. The reason for this is since all of the custom columns
are custom `randomize.Struct` can't construct randomized instances of them. The solution for this is implementing
bespoke randomization procedures like `c.Segment = custom_types.SegmentRandom()`. Hence any custom type
must implement a method like `custom_types.<identifier>Random`. Note that `Null<identifier>` types must implement
`custom_types.Null<identifier>Random` methods.

Similarly

```go
seed := randomize.NewSeed()
var err error
{{$varNameSingular}} := &{{$tableNameSingular}}{}
if err = randomize.Struct(seed, {{$varNameSingular}}, {{$varNameSingular}}DBTypes, true, {{$varNameSingular}}ColumnsWithDefault...); err != nil {
    t.Errorf("Unable to randomize {{$tableNameSingular}} struct: %s", err)
}
```

is replaced with

```go
{{template "isCustomSimple" .}}
```
which is a helper template @ `databrary/config/sqlboiler/templates/templates_helpers`

```go
{{define "isCustomSimple"}}
    {{- $hasCustom := .Table.HasCustom -}}
    {{- $varNameSingular := .Table.Name | singular | camelCase -}}
    {{- $tableNameSingular := .Table.Name | singular | titleCase -}}
    var err error
    seed := randomize.NewSeed()
    {{$varNameSingular}} := &{{$tableNameSingular}}{}
    {{- if not $hasCustom }}
    {{template "nonCustomRandom" . }}
    {{else}}
    {{template "customRandom" . }}
    {{template "customRandomRangeOne" . }}
    {{end}}
{{end}}
```

The distinguishing branch is the `{{else}}`, which generates code from these two templates

```go
{{define "customRandom"}}
    {{- $varNameSingular := .Table.Name | singular | camelCase -}}
    {{- $tableNameSingular := .Table.Name | singular | titleCase -}}
    if err = randomize.Struct(seed, {{$varNameSingular}}, {{$varNameSingular}}DBTypes, true, {{$varNameSingular}}ColumnsWithCustom...); err != nil {
        t.Errorf("Unable to randomize {{$tableNameSingular}} struct: %s", err)
    }
{{end}}
{{define "customRandomRangeOne"}}
    {{- $varNameSingular := .Table.Name | singular | camelCase -}}
    {{range $i, $v := .Table.GetCustomColumns -}}
    {{$varNameSingular}}.{{$v.Name | titleCase}} = {{$v.Type}}Random()
    {{end -}}
{{end}}
```

The easiest way to understand what a template does is to inspect the code they generate (which you can find by grepping for the parts of the templates that aren't
filled, such as `t.Errorf("Unable to randomize`). So these templates
generate

```go
if err = randomize.Struct(seed, account, accountDBTypes, true); err != nil {
    t.Errorf("Unable to randomize Account struct: %s", err)
}
```
and
```go
var err error
seed := randomize.NewSeed()
accountOne := &Account{}
accountTwo := &Account{}
if err = randomize.Struct(seed, accountOne, accountDBTypes, false, accountColumnsWithDefault...); err != nil {
    t.Errorf("Unable to randomize Account struct: %s", err)
}
if err = randomize.Struct(seed, accountTwo, accountDBTypes, false, accountColumnsWithDefault...); err != nil {
    t.Errorf("Unable to randomize Account struct: %s", err)
}
```

#### Imports

Any model that includes a custom type must import the package `"github.com/databrary/databrary/db/models/custom_types"`, along several others

```go
import (
    "bytes"
    "database/sql"
    "fmt"
    "reflect"
    "strings"
    "sync"
    "time"

    "github.com/databrary/databrary/db/models/custom_types"
    "github.com/pkg/errors"
    "github.com/vattle/sqlboiler/boil"
    "github.com/vattle/sqlboiler/queries"
    "github.com/vattle/sqlboiler/queries/qm"
    "github.com/vattle/sqlboiler/strmangle"
    "gopkg.in/nullbio/null.v6"
)
```

In SQLBoiler this is handled using a bunch of complicated logic for when to include
third party imports, SQLBoiler types, etc. in the preample to the source. My solution was to simply
use a regex to search for package references e.g. `custom_types.NullRelease` after the source has been written:

```go
func writeImports(top, out *bytes.Buffer) {
    top.WriteString("import (\n")
    for k, v := range imps {
        rgxType := regexp.MustCompile(fmt.Sprintf(`\b%s\b`, k))
        if rgxType.Match(out.Bytes()) {
            _, _ = fmt.Fprintf(top, "\t%s\n", v)
        }
    }
    top.WriteString(")\n")
}
```

Note this means that there should be no unqualified imports in any of the models or tests - there aren't any now and the templates shouldn't be adjusted
to use any. In fact you should probably never use unqualified imports ever for the rest of your life :)

### Views

Views don't have primary key and so many of the methods that are automatically generated for models that correspond to tables that do have primary keys
shouldn't be generated. For example

`func Find<identifier>(exec boil.Executor, id int, selectCols ...string) (*Account, error)`

shouldn't be generated. Hence `14_find.tpl` has a "guard" at the top of the template

`{{ if .Table.HasPrimaryKey }}` that prevents the generation of the `Find` method from being generated. `HasPrimaryKey` is a `Table` method

```go
func (t Table) HasPrimaryKey() bool {
    return t.PKey != nil
}
```

## Tests

All of the test harness code is template generated too, @ `databrary/config/sqlboiler/templates/templates_test/singleton` and `databrary/config/sqlboiler/templates/templates_test/main_test`.

**If you want to change test behavior do not forget to update the test generation templates!**

This is slightly counter-intuitive since it seems like at least the database connections should be boiler plate. Hence the only template that has actual template code in it is
`boil_suites_test.tpl`. `boil_suites_test.tpl` was extended to include a "live" test:

```go
func TestLive(t *testing.T) {
    if err := dbMain.setupLiveTest(); err != nil {
        fmt.Println("Unable to execute setupLiveTest:", err)
        os.Exit(-4)
    }
    var err error
    dbMain.liveTestDbConn, err = dbMain.conn(dbMain.LiveTestDBName)
    if err != nil {
        fmt.Println("failed to get test connection:", err)
    }
    dbMain.liveDbConn, err = dbMain.conn(dbMain.DbName)
    if err != nil {
        fmt.Println("failed to get live connection:", err)
    }
  {{ range $index, $table := .Tables}}
  {{- if not $table.IsJoinTable -}}
  {{- $tableName := $table.Name | plural | titleCase -}}
    {{ if $table.HasPrimaryKey }}
  t.Run("{{$tableName}}", test{{$tableName}}Live)
  {{end}}
  {{end -}}
  {{end}}
}
```

The live test (@ `databrary/config/sqlboiler/templates/templates_test/all.tpl`) selects all rows in a table, inserts them into a table on a test database with the same schema, dumps both tables, and then diffs them.
Again this test is not created for views. Despite being boilerplate `main_test.tpl`, `postgres_main.tpl`, `boil_main_test.tpl` have also been modified. The most significant changes are the
creation of a test database for use with the "live" tests:

```go
func (p *pgTester) setupLiveTest() error {
    var err error

    if err = p.dropDB(p.LiveTestDBName); err != nil {
        return err
    }
    if err = p.createDB(p.LiveTestDBName); err != nil {
        return err
    }

    dumpCmd := exec.Command("pg_dump", "--schema-only", p.DbName)
    dumpCmd.Env = append(os.Environ(), p.pgEnv()...)
    createCmd := exec.Command("psql", p.LiveTestDBName)
    createCmd.Env = append(os.Environ(), p.pgEnv()...)

    r, w := io.Pipe()
    dumpCmd.Stdout = w
    conDestroyer := newFKeyDestroyer(rgxPGCon, r)
    chkReplacer := newConReplacer(rgxCheckCon, conDestroyer)
    trigDestroyer := newFKeyDestroyer(rgxPGTrig, chkReplacer)
    createCmd.Stdin = trigDestroyer

    ...

}
```

The 3 lines

```go
conDestroyer := newFKeyDestroyer(rgxPGCon, r)
chkReplacer := newConReplacer(rgxCheckCon, conDestroyer)
trigDestroyer := newFKeyDestroyer(rgxPGTrig, chkReplacer)
```

remove constraints in the schema in order to make certain tests possible (foreign key constraints are dropped, checks are dropped, and triggers are dropped).
The last two are necessary for the live tests. Despite this there will be exclusion failures in tests such as

`
container_test.go:898: models: unable to insert into slot_release: pq: conflicting key value violates exclusion constraint "slot_release_overlap_excl"
`

I considered dropping these exclusion constraints from the test database but this caused other tests to fail (don't remember which).

