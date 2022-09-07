package filestore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	esh "github.com/go-ee/utils/eh"
	"github.com/go-ee/utils/eio"
	"os"
	"path/filepath"
	"reflect"
	"sync"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	eh "github.com/looplab/eventhorizon"
)

// ErrModelNotSet is when an model factory is not set on the Repo.
var ErrModelNotSet = errors.New("model not set")

// Repo implements an in memory repository of read models.
type Repo struct {
	*Base
	// The outer map is with namespace as key, the inner with aggregate ID.
	db   map[string]map[uuid.UUID]eh.Entity
	dbMu sync.RWMutex

	// A list of all item ids, only the order is used.
	// The outer map is for the namespace.
	ids        map[string][]uuid.UUID
	factoryFn  func() eh.Entity
	entityType reflect.Type
}

// NewRepo creates a new Repo.
func NewRepo(folder string) (ret *Repo, err error) {
	if err = os.MkdirAll(folder, DEFAULT_FOLDER_PERM); err == nil {
		ret = &Repo{
			Base: NewBase(folder),
			ids:  map[string][]uuid.UUID{},
			db:   map[string]map[uuid.UUID]eh.Entity{},
		}
	}
	return
}

func (r *Repo) InnerRepo(context.Context) eh.ReadRepo {
	return nil
}

// Find implements the Find method of the eventhorizon.ReadRepo interface.
func (r *Repo) Find(ctx context.Context, id uuid.UUID) (ret eh.Entity, err error) {
	if r.factoryFn == nil {
		return nil, &eh.RepoError{
			Err:      fmt.Errorf("%v: %v", ErrModelNotSet, esh.ContextGetNamespace(ctx)),
			Op:       eh.RepoOpFind,
			EntityID: id,
		}
	}

	r.dbMu.RLock()
	defer r.dbMu.RUnlock()

	var ns string
	if ns, err = r.namespace(ctx); err != nil {
		return
	}

	item, ok := r.db[ns][id]
	if !ok {
		return nil, &eh.RepoError{
			Err:      fmt.Errorf("%v: %v", eh.ErrEntityNotFound, esh.ContextGetNamespace(ctx)),
			Op:       eh.RepoOpFind,
			EntityID: id,
		}
	}
	ret = r.factoryFn()
	copier.Copy(ret, item)

	return
}

// FindAll implements the FindAll method of the eventhorizon.ReadRepo interface.
func (r *Repo) FindAll(ctx context.Context) (ret []eh.Entity, err error) {
	if r.factoryFn == nil {
		return nil, &eh.RepoError{
			Err: fmt.Errorf("%v: %v", ErrModelNotSet, esh.ContextGetNamespace(ctx)),
			Op:  eh.RepoOpFindAll,
		}
	}

	r.dbMu.RLock()
	defer r.dbMu.RUnlock()

	var ns string
	if ns, err = r.namespace(ctx); err != nil {
		return
	}

	ret = []eh.Entity{}
	for _, id := range r.ids[ns] {
		if item, ok := r.db[ns][id]; ok {
			entity := r.factoryFn()
			copier.Copy(entity, item)
			ret = append(ret, entity)
		}
	}
	return
}

// Save implements the Save method of the eventhorizon.WriteRepo interface.
func (r *Repo) Save(ctx context.Context, entity eh.Entity) (err error) {
	if r.factoryFn == nil {
		return &eh.RepoError{
			Err:      fmt.Errorf("%v: %v", ErrModelNotSet, esh.ContextGetNamespace(ctx)),
			Op:       eh.RepoOpSave,
			EntityID: entity.EntityID(),
		}
	}

	r.dbMu.Lock()
	defer r.dbMu.Unlock()

	var ns string
	if ns, err = r.namespace(ctx); err != nil {
		return
	}

	if entity.EntityID() == uuid.Nil {
		return &eh.RepoError{
			Err: fmt.Errorf("could not save entity, missing entity id: %v", esh.ContextGetNamespace(ctx)),
			Op:  eh.RepoOpSave,
		}
	}

	id := entity.EntityID()
	if _, ok := r.db[ns][id]; !ok {
		r.ids[ns] = append(r.ids[ns], id)
	}
	toInsert := r.factoryFn()
	copier.Copy(toInsert, entity)
	r.db[ns][id] = toInsert

	err = r.saveFile(ns)

	return
}

// Remove implements the Remove method of the eventhorizon.WriteRepo interface.
func (r *Repo) Remove(ctx context.Context, id uuid.UUID) (err error) {
	var ns string
	if ns, err = r.namespace(ctx); err != nil {
		return
	}

	r.dbMu.Lock()
	defer r.dbMu.Unlock()

	if _, ok := r.db[ns][id]; ok {
		delete(r.db[ns], id)

		index := -1
		for i, d := range r.ids[ns] {
			if id == d {
				index = i
				break
			}
		}
		r.ids[ns] = append(r.ids[ns][:index], r.ids[ns][index+1:]...)

		r.saveFile(ns)
		return
	}

	err = &eh.RepoError{
		Err: fmt.Errorf("%v: %v", eh.ErrEntityNotFound, esh.ContextGetNamespace(ctx)),
		Op:  eh.RepoOpRemove,
	}
	return
}

// SetEntityFactory sets a factory function that creates concrete entity types.
func (r *Repo) SetEntityFactory(f func() eh.Entity) {
	r.factoryFn = f
	r.entityType = reflect.TypeOf(f())
}

// Helper to get the namespace and ensure that its data exists.
func (r *Repo) namespace(ctx context.Context) (ns string, err error) {
	ns = esh.ContextGetNamespace(ctx)
	err = r.loadFile(ns)
	return
}

func (r *Repo) loadFile(ns string) (err error) {
	if _, ok := r.db[ns]; !ok {
		var fileJson string
		if fileJson, err = r.buildFileNameAndMkdirParents(ns); err != nil {
			return
		}
		if items, jsonErr := eio.LoadArrayJsonByReflect(fileJson, r.entityType); err != nil {
			err = &eh.RepoError{
				Err: fmt.Errorf("could not load entity, %v: %v", jsonErr, ns),
				Op:  eh.RepoOpFind,
			}
		} else {
			data := map[uuid.UUID]eh.Entity{}
			ids := make([]uuid.UUID, len(items))
			for i, item := range items {
				entity := item.(eh.Entity)
				data[entity.EntityID()] = entity
				ids[i] = entity.EntityID()
			}
			r.db[ns] = data
			r.ids[ns] = ids
		}
	}
	return
}

func (r *Repo) saveFile(ns string) (err error) {
	db := r.db[ns]
	ids := r.ids[ns]
	items := make([]eh.Entity, len(ids))
	for i, id := range ids {
		items[i] = db[id]
	}

	data, _ := json.MarshalIndent(items, "", "  ")
	var fileJson string
	if fileJson, err = r.buildFileNameAndMkdirParents(ns); err != nil {
		return
	}

	if writeErr := os.WriteFile(fileJson, data, r.defaultFilePerm); writeErr != nil {
		err = &eh.RepoError{
			Err: fmt.Errorf("could not save entity:  %v", ns),
			Op:  eh.RepoOpSave,
		}
	}
	return
}

func (r *Repo) buildFileNameAndMkdirParents(ns string) (ret string, err error) {
	ret = filepath.Join(r.folder, ns+".json")
	err = r.MkdirParents(ret)
	return
}

func (r *Repo) MkdirParents(file string) error {
	return os.MkdirAll(filepath.Dir(file), r.defaultFolderPerm)
}

func (r *Repo) Close() (err error) {
	//noting todo for now
	return
}

/*

type ReadWriteRepoDelegate struct {
	Factory func() (ret eventhorizon.ReadWriteRepo, err error)
	repo    eventhorizon.ReadWriteRepo
}

func (o *ReadWriteRepoDelegate) delegate() (ret eventhorizon.ReadWriteRepo, err error) {
	if o.repo == nil {
		o.repo, err = o.Factory()
	}
	ret = o.repo
	return
}

func (o *ReadWriteRepoDelegate) Save(ctx context.Context, entity eventhorizon.Entity) (err error) {
	var repo eventhorizon.ReadWriteRepo
	if repo, err = o.delegate(); err == nil {
		err = repo.Save(ctx, entity)
	}
	return
}

func (o *ReadWriteRepoDelegate) Remove(ctx context.Context, id uuid.UUID) (err error) {
	var repo eventhorizon.ReadWriteRepo
	if repo, err = o.delegate(); err == nil {
		err = repo.Remove(ctx, id)
	}
	return
}

func (o *ReadWriteRepoDelegate) Parent() (ret eventhorizon.ReadRepo) {
	if repo, err := o.delegate(); err == nil {
		ret = repo.Parent()
	}
	return
}

func (o *ReadWriteRepoDelegate) Find(ctx context.Context, id uuid.UUID) (ret eventhorizon.Entity, err error) {
	var repo eventhorizon.ReadWriteRepo
	if repo, err = o.delegate(); err == nil {
		ret, err = repo.Find(ctx, id)
	}
	return
}

func (o *ReadWriteRepoDelegate) FindAll(ctx context.Context) (ret []eventhorizon.Entity, err error) {
	var repo eventhorizon.ReadWriteRepo
	if repo, err = o.delegate(); err == nil {
		if ret, err = repo.FindAll(ctx); err == nil {
			ret = o.FilterDeleted(ctx, ret)
		}
	}
	return
}

func (o *ReadWriteRepoDelegate) FilterDeleted(ctx context.Context, ret []eventhorizon.Entity) []eventhorizon.Entity {
	n := 0
	for _, x := range ret {
		if e, ok := x.(Entity); ok {
			if e.Deleted() == nil {
				ret[n] = x
				n++
			} else {
				o.repo.Remove(ctx, e.EntityID())
			}
		} else {
			ret[n] = x
			n++
		}
	}
	ret = ret[:n]
	return ret
}
*/
