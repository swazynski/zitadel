import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';

import { InfoOverlayComponent } from './info-overlay.component';

@NgModule({
  declarations: [InfoOverlayComponent],
  imports: [CommonModule, MatButtonModule, MatIconModule],
  exports: [InfoOverlayComponent],
})
export class InfoOverlayModule {}
