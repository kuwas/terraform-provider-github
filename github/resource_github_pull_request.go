package github

import (
	"context"
	"fmt"
	"github.com/google/go-github/v31/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
	"strconv"
)

func resourceGithubPullRequest() *schema.Resource {
	return &schema.Resource{
		Read:   resourceGithubPullRequestRead,
		Create: resourceGithubPullRequestCreate,
		Update: resourceGithubPullRequestUpdate,
		Delete: resourceGithubPullRequestDelete,
		Schema: map[string]*schema.Schema{
			"uid":        {Type: schema.TypeString, Required: true, ForceNew: true},
			"repository": {Type: schema.TypeString, Required: true, ForceNew: true},
			"base":       {Type: schema.TypeString, Required: true},
			"head":       {Type: schema.TypeString, Required: true},
			"title":      {Type: schema.TypeString, Required: true},
			"body":       {Type: schema.TypeString, Required: true},
		},
	}
}

func resourceGithubPullRequestRead(d *schema.ResourceData, m interface{}) error {

	ctx := context.Background()
	owner := m.(*Owner).name
	client := m.(*Owner).v3client

	id := d.Id()
	repository, number, _ := parseTwoPartID(id, "repository", "number")
	num, _ := strconv.Atoi(number)
	log.Printf("[DEBUG] Reading GitHub Pull Request in repository %s/%s", repository, number)
	pr, _, err := client.PullRequests.Get(ctx, owner, repository, num)
	if err != nil {
		return fmt.Errorf("Error reading GitHub Pull Request in repository %s/%s: %s", repository, number, err)
	}

	if !*pr.Merged && *pr.State == "closed" {
		d.Set("uid", "")
	}

	return nil

}

func resourceGithubPullRequestUpdate(d *schema.ResourceData, m interface{}) error { return resourceGithubPullRequestCreate(d, m) }
func resourceGithubPullRequestCreate(d *schema.ResourceData, m interface{}) error {

	ctx := context.Background()
	owner := m.(*Owner).name
	client := m.(*Owner).v3client

	repository := d.Get("repository").(string)
	base := d.Get("base").(string)
	head := d.Get("head").(string)
	title := d.Get("title").(string)
	body := d.Get("body").(string)

	log.Printf("[DEBUG] Creating GitHub Pull Request in repository %s from %s to %s", repository, base, head)
	pr, _, err := client.PullRequests.Create(ctx, owner, repository, &github.NewPullRequest{
		Base:  github.String(base),
		Head:  github.String(head),
		Title: github.String(title),
		Body:  github.String(body),
	})
	if err != nil {
		return fmt.Errorf("Error creating GitHub Pull Request in repository %s from %s to %s: %s", repository, base, head, err)
	}

	d.SetId(buildTwoPartID(repository, strconv.Itoa(*pr.Number)))
	return nil

}

func resourceGithubPullRequestDelete(d *schema.ResourceData, m interface{}) error {

	ctx := context.Background()
	owner := m.(*Owner).name
	client := m.(*Owner).v3client

	id := d.Id()
	repository, number, _ := parseTwoPartID(id, "repository", "number")
	num, _ := strconv.Atoi(number)
	log.Printf("[DEBUG] Deleting GitHub Pull Request in repository %s/%s", repository, number)
	_, _, err := client.PullRequests.Edit(ctx, owner, repository, num, &github.PullRequest{State: github.String("closed")})
	if err != nil {
		return fmt.Errorf("Error deleting GitHub Pull Request in repository %s/%s: %s", repository, number, err)
	}

	d.SetId("")
	return nil

}
