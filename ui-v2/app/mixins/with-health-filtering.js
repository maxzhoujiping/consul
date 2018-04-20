import Mixin from '@ember/object/mixin';
import WithFiltering from 'consul-ui/mixins/with-filtering';
import { computed, get } from '@ember/object';
import ucfirst from 'consul-ui/utils/ucfirst';
import humanize from 'consul-ui/utils/humanize';

const countStatus = function(items, status) {
  if (status === '') {
    return get(items, 'length');
  }
  const key = `Checks${ucfirst(status)}`;
  return items.reduce(function(prev, item, i, arr) {
    const num = get(item, key);
    return (
      prev +
        (typeof num !== 'undefined'
          ? num
          : get(item, 'Checks').filter(function(item) {
              return item.Status === status;
            }).length) || 0
    );
  }, 0);
};
export default Mixin.create(WithFiltering, {
  filters: {
    status: '',
  },
  healthFilters: computed('items', function() {
    const items = get(this, 'items');
    return ['', 'passing', 'warning', 'critical'].map(function(item) {
      return {
        label: `${item === '' ? 'All' : ucfirst(item)} (${humanize(countStatus(items, item))})`,
        value: item,
      };
    });
  }),
});