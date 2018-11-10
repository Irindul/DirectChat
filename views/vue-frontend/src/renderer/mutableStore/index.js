const store = {
  state: {},
  push: (key) => {
    store.state = {
      ...store.state,
      ...key,
    }
  },
  clean: () => {
    store.state = {};
  }
};

export default store;