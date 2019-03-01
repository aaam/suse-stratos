
import { pairwise } from 'rxjs/operators';
import { Injectable } from '@angular/core';
import { Store } from '@ngrx/store';

import { ITableColumn } from '../../list-table/table.types';
import { IListConfig, ListViewTypes } from '../../list.component.types';
import { endpointColumns } from '../endpoint/endpoints-list-config.service';
import { EndpointModel, endpointStoreNames } from '../../../../../../../store/src/types/endpoint.types';
import { selectUpdateInfo } from '../../../../../../../store/src/selectors/api.selectors';
import { AppState } from '../../../../../../../store/src/app-state';
import {
  BaseEndpointsDataSource
} from './base-endpoints-data-source';
import { CfEndpointCardComponent } from './cf-endpoint-card/cf-endpoint-card.component';

@Injectable()
export class CFEndpointsListConfigService implements IListConfig<EndpointModel> {
  columns: ITableColumn<EndpointModel>[];
  isLocal = true;
  dataSource: BaseEndpointsDataSource;
  viewType = ListViewTypes.CARD_ONLY;
  cardComponent = CfEndpointCardComponent;
  text = {
    title: '',
    filter: 'Filter Endpoints',
    noEntries: 'There are no endpoints'
  };
  enableTextFilter = true;
  tableFixedRowHeight = true;

  private handleAction(item, effectKey, handleChange) {
    const disSub = this.store.select(selectUpdateInfo(
      endpointStoreNames.type,
      item.guid,
      effectKey,
    )).pipe(
      pairwise())
      .subscribe(([oldVal, newVal]) => {
        if (!newVal.error && (oldVal.busy && !newVal.busy)) {
          handleChange([oldVal, newVal]);
          disSub.unsubscribe();
        }
      });
  }

  constructor(
    private store: Store<AppState>
  ) {
    this.columns = endpointColumns.filter(column => {
      return column.columnId !== 'type';
    });
    this.dataSource = new BaseEndpointsDataSource(this.store, this, 'cf');
  }
  public getColumns = () => this.columns;
  public getGlobalActions = () => [];
  public getMultiActions = () => [];
  public getSingleActions = () => [];
  public getMultiFiltersConfigs = () => [];
  public getDataSource = () => this.dataSource;
}
