package fingprints

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type Plugin interface {
    Fingerprint(body string, headers map[string][]string) bool
    Name() string
}
