import {Injectable} from '@angular/core';
import {ActivatedRouteSnapshot, CanActivate, Router, RouterStateSnapshot} from '@angular/router';

import {AuthService} from '@services/auth.service';

@Injectable({providedIn: 'root'})
export class AuthChecker implements CanActivate {
    constructor(
        private router: Router,
        private authenticationService: AuthService
    ) {
    }

    canActivate(route: ActivatedRouteSnapshot, state: RouterStateSnapshot) {
        // logged in, return
        if (!!this.authenticationService.currentSessionValue) {
            return true;
        }
        // not logged in so redirect to login page with the return url
        this.router.navigate(['/login'], !!state.url && state.url !== '/' ? {queryParams: {returnUrl: state.url}} : {});
        return false;
    }
}

@Injectable({providedIn: 'root'})
export class SignChecker implements CanActivate {
    constructor(private authenticationService: AuthService) {
    }

    canActivate(route: ActivatedRouteSnapshot, state: RouterStateSnapshot) {
        return (state.url.indexOf('/registration') === 0 || state.url.indexOf('/login') === 0)
            && !this.authenticationService.currentSessionValue;
    }
}

