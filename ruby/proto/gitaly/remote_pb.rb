# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: remote.proto

require 'google/protobuf'

require 'lint_pb'
require 'shared_pb'
Google::Protobuf::DescriptorPool.generated_pool.build do
  add_file("remote.proto", :syntax => :proto3) do
    add_message "gitaly.AddRemoteRequest" do
      optional :repository, :message, 1, "gitaly.Repository"
      optional :name, :string, 2
      optional :url, :string, 3
      repeated :mirror_refmaps, :string, 5
    end
    add_message "gitaly.AddRemoteResponse" do
    end
    add_message "gitaly.RemoveRemoteRequest" do
      optional :repository, :message, 1, "gitaly.Repository"
      optional :name, :string, 2
    end
    add_message "gitaly.RemoveRemoteResponse" do
      optional :result, :bool, 1
    end
    add_message "gitaly.FetchInternalRemoteRequest" do
      optional :repository, :message, 1, "gitaly.Repository"
      optional :remote_repository, :message, 2, "gitaly.Repository"
    end
    add_message "gitaly.FetchInternalRemoteResponse" do
      optional :result, :bool, 1
    end
    add_message "gitaly.UpdateRemoteMirrorRequest" do
      optional :repository, :message, 1, "gitaly.Repository"
      optional :ref_name, :string, 2
      repeated :only_branches_matching, :bytes, 3
      optional :ssh_key, :string, 4
      optional :known_hosts, :string, 5
      optional :keep_divergent_refs, :bool, 6
    end
    add_message "gitaly.UpdateRemoteMirrorResponse" do
      repeated :divergent_refs, :bytes, 1
    end
    add_message "gitaly.FindRemoteRepositoryRequest" do
      optional :remote, :string, 1
      optional :storage_name, :string, 2
    end
    add_message "gitaly.FindRemoteRepositoryResponse" do
      optional :exists, :bool, 1
    end
    add_message "gitaly.FindRemoteRootRefRequest" do
      optional :repository, :message, 1, "gitaly.Repository"
      optional :remote, :string, 2
    end
    add_message "gitaly.FindRemoteRootRefResponse" do
      optional :ref, :string, 1
    end
    add_message "gitaly.ListRemotesRequest" do
      optional :repository, :message, 1, "gitaly.Repository"
    end
    add_message "gitaly.ListRemotesResponse" do
      repeated :remotes, :message, 1, "gitaly.ListRemotesResponse.Remote"
    end
    add_message "gitaly.ListRemotesResponse.Remote" do
      optional :name, :string, 1
      optional :fetch_url, :string, 2
      optional :push_url, :string, 3
    end
    add_message "gitaly.FetchRemoteWithStatusRequest" do
      optional :repository, :message, 1, "gitaly.Repository"
      optional :remote, :string, 2
      optional :force, :bool, 3
      optional :no_tags, :bool, 4
      optional :timeout, :int32, 5
      optional :ssh_key, :string, 6
      optional :known_hosts, :string, 7
      optional :no_prune, :bool, 9
      optional :remote_params, :message, 10, "gitaly.Remote"
    end
    add_message "gitaly.FetchRemoteWithStatusResponse" do
      repeated :ref_updates, :message, 1, "gitaly.FetchRemoteWithStatusResponse.RefUpdate"
    end
    add_message "gitaly.FetchRemoteWithStatusResponse.RefUpdate" do
      optional :status, :enum, 1, "gitaly.FetchRemoteWithStatusResponse.Status"
      optional :ref, :string, 2
    end
    add_enum "gitaly.FetchRemoteWithStatusResponse.Status" do
      value :FAST_FORWARD_UPDATE, 0
      value :FORCED_UPDATE, 1
      value :PRUNED, 2
      value :TAG_UPDATE, 3
      value :FETCHED, 4
      value :UPDATE_FAILED, 5
      value :UNCHANGED, 6
    end
  end
end

module Gitaly
  AddRemoteRequest = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.AddRemoteRequest").msgclass
  AddRemoteResponse = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.AddRemoteResponse").msgclass
  RemoveRemoteRequest = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.RemoveRemoteRequest").msgclass
  RemoveRemoteResponse = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.RemoveRemoteResponse").msgclass
  FetchInternalRemoteRequest = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.FetchInternalRemoteRequest").msgclass
  FetchInternalRemoteResponse = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.FetchInternalRemoteResponse").msgclass
  UpdateRemoteMirrorRequest = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UpdateRemoteMirrorRequest").msgclass
  UpdateRemoteMirrorResponse = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UpdateRemoteMirrorResponse").msgclass
  FindRemoteRepositoryRequest = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.FindRemoteRepositoryRequest").msgclass
  FindRemoteRepositoryResponse = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.FindRemoteRepositoryResponse").msgclass
  FindRemoteRootRefRequest = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.FindRemoteRootRefRequest").msgclass
  FindRemoteRootRefResponse = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.FindRemoteRootRefResponse").msgclass
  ListRemotesRequest = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.ListRemotesRequest").msgclass
  ListRemotesResponse = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.ListRemotesResponse").msgclass
  ListRemotesResponse::Remote = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.ListRemotesResponse.Remote").msgclass
  FetchRemoteWithStatusRequest = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.FetchRemoteWithStatusRequest").msgclass
  FetchRemoteWithStatusResponse = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.FetchRemoteWithStatusResponse").msgclass
  FetchRemoteWithStatusResponse::RefUpdate = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.FetchRemoteWithStatusResponse.RefUpdate").msgclass
  FetchRemoteWithStatusResponse::Status = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.FetchRemoteWithStatusResponse.Status").enummodule
end
