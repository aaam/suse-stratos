import { Component, OnInit, Input } from '@angular/core';

@Component({
  selector: 'app-resource-alert-view',
  templateUrl: './resource-alert-view.component.html',
  styleUrls: ['./resource-alert-view.component.scss']
})
export class ResourceAlertViewComponent implements OnInit {

  alertInfo;

  @Input()
  set alerts(data: any) {
    if (data) {
      const alerts = data.alerts ? data.alerts : data;
      this.alertInfo = this.normalize(alerts);
    }
  }

  @Input() showHeader = true;

  constructor() { }

  ngOnInit() { }

  normalize(data) {
    // Normalize the alerts into groups
    const normalized = {};
    data.forEach(item => {
      const path = item.namespace ? `${item.namespace}/${item.name}` : item.name;
      if (!normalized[path]) {
        normalized[path] = [];
      }
      item.path = path;
      normalized[path].push(item);
    });

    const arr = [];
    Object.keys(normalized).forEach(group => {
      arr.push({
        name: group,
        alerts: normalized[group]
      });
    });
    return arr;
  }

}
