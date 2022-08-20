// The feat package contains several sub-folders each implementing a dedicated feature.
package feat

type Feature interface {
	Start() error
	Close() error
}
