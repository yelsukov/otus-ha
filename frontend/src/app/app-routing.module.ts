import {NgModule} from '@angular/core';

import {RouterModule, Routes} from '@angular/router';

import {DashboardComponent} from './components/dashboard/dashboard.component';
import {LoginComponent} from './components/login/login.component';
import {RegistrationComponent} from './components/registration/registration.component';
import {FriendsComponent} from './components/friends/friends.component';
import {ProfileComponent} from './components/profile/profile.component';
import {UserComponent} from './components/user/user.component';
import {PageNotFoundComponent} from './components/page-not-found/page-not-found.component';
import {AuthChecker, SignChecker} from './helpers/AuthChecker';

const routes: Routes = [
    {path: '', component: DashboardComponent, canActivate: [AuthChecker]},
    {path: 'login', component: LoginComponent, canActivate: [SignChecker]},
    {path: 'registration', component: RegistrationComponent, canActivate: [SignChecker]},
    {path: 'friends', component: FriendsComponent, canActivate: [AuthChecker]},
    {path: 'profile', component: ProfileComponent, canActivate: [AuthChecker]},
    {path: 'user/:id', component: UserComponent, canActivate: [AuthChecker]},

    {path: '**', component: PageNotFoundComponent}
];


@NgModule({
    imports: [RouterModule.forRoot(routes)],
    exports: [RouterModule]
})
export class AppRoutingModule {
}
