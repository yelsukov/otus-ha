import {Component} from '@angular/core';
import {Router} from '@angular/router';

import {AuthService} from '@services/auth.service';

@Component({selector: 'app', templateUrl: 'app.component.html'})
export class AppComponent {
    currentUser: any;

    constructor(
        private router: Router,
        private authenticationService: AuthService
    ) {
        this.authenticationService.currentSession.subscribe(x => this.currentUser = !!x ? x : null);
    }

    logout() {
        this.authenticationService.logout();
    }
}
