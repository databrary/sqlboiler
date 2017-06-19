package boilingcore

var imps = map[string]string{

	"null.Float32": `"gopkg.in/nullbio/null.v6"`,

	"null.Float64": `"gopkg.in/nullbio/null.v6"`,

	"null.Int": `"gopkg.in/nullbio/null.v6"`,

	"null.Int8": `"gopkg.in/nullbio/null.v6"`,

	"null.Int16": `"gopkg.in/nullbio/null.v6"`,

	"null.Int32": `"gopkg.in/nullbio/null.v6"`,

	"null.Int64": `"gopkg.in/nullbio/null.v6"`,

	"null.Uint": `"gopkg.in/nullbio/null.v6"`,

	"null.Uint8": `"gopkg.in/nullbio/null.v6"`,

	"null.Uint16": `"gopkg.in/nullbio/null.v6"`,

	"null.Uint32": `"gopkg.in/nullbio/null.v6"`,

	"null.Uint64": `"gopkg.in/nullbio/null.v6"`,

	"null.String": `"gopkg.in/nullbio/null.v6"`,

	"null.Bool": `"gopkg.in/nullbio/null.v6"`,

	"null.Time": `"gopkg.in/nullbio/null.v6"`,

	"null.JSON": `"gopkg.in/nullbio/null.v6"`,

	"null.Bytes": `"gopkg.in/nullbio/null.v6"`,

	`time\.`: `"time"`,

	"types.JSON": `"github.com/databrary/sqlboiler/types"`,

	"types.BytesArray": `"github.com/databrary/sqlboiler/types"`,

	"types.Int64Array": `"github.com/databrary/sqlboiler/types"`,

	"types.Float64Array": `"github.com/databrary/sqlboiler/types"`,

	"types.BoolArray": `"github.com/databrary/sqlboiler/types"`,

	"types.StringArray": `"github.com/databrary/sqlboiler/types"`,

	"types.Hstore": `"github.com/databrary/sqlboiler/types"`,

	"custom_types.Interval": `"github.com/databrary/databrary/db/models/custom_types"`,

	"custom_types.NullInterval": `"github.com/databrary/databrary/db/models/custom_types"`,

	"custom_types.Inet": `"github.com/databrary/databrary/db/models/custom_types"`,

	"custom_types.NullInet": `"github.com/databrary/databrary/db/models/custom_types"`,

	"custom_types.Segment": `"github.com/databrary/databrary/db/models/custom_types"`,

	"custom_types.NullSegment": `"github.com/databrary/databrary/db/models/custom_types"`,

	"custom_types.Release": `"github.com/databrary/databrary/db/models/custom_types"`,

	"custom_types.NullRelease": `"github.com/databrary/databrary/db/models/custom_types"`,

	"custom_types.Permission": `"github.com/databrary/databrary/db/models/custom_types"`,

	"custom_types.NullPermission": `"github.com/databrary/databrary/db/models/custom_types"`,

	"custom_types.NoticeDelivery": `"github.com/databrary/databrary/db/models/custom_types"`,

	"custom_types.NullNoticeDelivery": `"github.com/databrary/databrary/db/models/custom_types"`,

	"custom_types.DataType": `"github.com/databrary/databrary/db/models/custom_types"`,

	"custom_types.NullDataType": `"github.com/databrary/databrary/db/models/custom_types"`,

	"reflect": `"reflect"`, //TODO: this was uncommented

	`sort\.`: `"sort"`,

	"difflib": `"github.com/pmezard/go-difflib/difflib"`,

	"queries": `"github.com/databrary/sqlboiler/queries"`,

	"qm": `"github.com/databrary/sqlboiler/queries/qm"`,

	"strmangle": `"github.com/databrary/sqlboiler/strmangle"`,

	"sync": `"sync"`,

	"flag": `"flag"`,

	"filepath": `"path/filepath"`,

	"vala": `"github.com/kat-co/vala"`,

	"rand": `"math/rand"`,

	"regexp": `"regexp"`,

	"boil": `"github.com/databrary/sqlboiler/boil"`,

	"testing": `"testing"`,

	"bytes": `"bytes"`,

	`sql\.`: `"database/sql"`,

	"fmt": `"fmt"`,

	"io": `"io"`,

	"ioutil": `"io/ioutil"`,

	"os": `"os"`,

	"exec.Command": `"os/exec"`,

	"strings": `"strings"`,

	"errors": `"github.com/pkg/errors"`,

	"viper": `"github.com/spf13/viper"`,

	"drivers": `"github.com/databrary/sqlboiler/bdb/drivers"`,

	"randomize": `"github.com/databrary/sqlboiler/randomize"`,
}

func removeDuplicates(dedup []string) []string {
	if len(dedup) <= 1 {
		return dedup
	}

	for i := 0; i < len(dedup)-1; i++ {
		for j := i + 1; j < len(dedup); j++ {
			if dedup[i] != dedup[j] {
				continue
			}

			if j != len(dedup)-1 {
				dedup[j] = dedup[len(dedup)-1]
				j--
			}
			dedup = dedup[:len(dedup)-1]
		}
	}

	return dedup
}
